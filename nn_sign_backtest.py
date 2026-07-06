#!/usr/bin/env python3
"""
NN Sign Backtest — Full Statistical Rigor
=========================================
- S&P 500 (1950–2026) and Gold (1975–2026)
- Multiple frequencies: daily, weekly, monthly, quarterly, period-level
- Bootstrap confidence intervals (10,000 resamples)
- Out-of-sample: train pre-2000, test post-2000
- Mann-Whitney U for period-level (respects autocorrelation)
- Effect sizes (Cohen's d)
- FDR correction across all 12 signs
"""

import json, math, sys
from datetime import datetime, timedelta
from collections import defaultdict
import yfinance as yf
import pandas as pd
import swisseph as swe
from scipy import stats
import numpy as np

# ── Configuration ────────────────────────────────────────────────────
ASSETS = {
    'SPX': ('^GSPC', '1950-01-01'),
    'GLD': ('GC=F', '1975-01-01'),
}

SIGNS = [
    'Aries', 'Taurus', 'Gemini', 'Cancer',
    'Leo', 'Virgo', 'Libra', 'Scorpio',
    'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces',
]

BOOTSTRAP_N = 10_000
OOS_SPLIT = datetime(2000, 1, 1)
RANDOM_SEED = 42

np.random.seed(RANDOM_SEED)

# ── NN Sign from longitude ───────────────────────────────────────────
def nn_sign(longitude):
    """Return zodiac sign name for a given ecliptic longitude (0-360)."""
    idx = int(longitude / 30) % 12
    return SIGNS[idx]

# ── Compute NN positions at 15-day intervals ─────────────────────────
def compute_nn_positions(start_date, end_date):
    """Return dict: date_str -> sign_name, sampled every 15 days."""
    positions = {}
    current = start_date
    while current <= end_date:
        jd = swe.julday(current.year, current.month, current.day, 12.0)
        result, _ = swe.calc_ut(jd, swe.MEAN_NODE, swe.FLG_SWIEPH)
        nn_lon = result[0]
        sign = nn_sign(nn_lon)
        positions[current.strftime('%Y-%m-%d')] = sign
        current += timedelta(days=15)
    return positions

# ── Map trading days to NN sign ──────────────────────────────────────
def map_dates_to_sign(trading_dates, nn_positions):
    """For each trading date, find the nearest prior NN sample."""
    nn_dates = sorted(nn_positions.keys())
    mapping = {}
    for td in trading_dates:
        td_str = td.strftime('%Y-%m-%d')
        # Find nearest prior NN sample
        best = None
        for nd in nn_dates:
            if nd <= td_str:
                best = nd
            else:
                break
        if best:
            mapping[td_str] = nn_positions[best]
    return mapping

# ── Bootstrap confidence interval ────────────────────────────────────
def bootstrap_ci(data, n_resamples=BOOTSTRAP_N, ci=95):
    """Return (lower, upper) bootstrap CI for the mean."""
    data = np.array(data)
    means = []
    for _ in range(n_resamples):
        sample = np.random.choice(data, size=len(data), replace=True)
        means.append(np.mean(sample))
    means = np.array(means)
    alpha = (100 - ci) / 2
    lower = np.percentile(means, alpha)
    upper = np.percentile(means, 100 - alpha)
    return lower, upper

# ── Cohen's d ────────────────────────────────────────────────────────
def cohens_d(group, baseline):
    """Cohen's d: (mean_group - mean_baseline) / pooled_std."""
    g = np.array(group)
    b = np.array(baseline)
    mean_diff = np.mean(g) - np.mean(b)
    n1, n2 = len(g), len(b)
    var1 = np.var(g, ddof=1)
    var2 = np.var(b, ddof=1)
    pooled_std = math.sqrt(((n1 - 1) * var1 + (n2 - 1) * var2) / (n1 + n2 - 2))
    if pooled_std == 0:
        return 0.0
    return mean_diff / pooled_std

# ── Annualized return from monthly returns ───────────────────────────
def annualize(monthly_returns):
    """Geometric annualization from monthly returns."""
    if not monthly_returns:
        return 0.0
    product = 1.0
    for r in monthly_returns:
        product *= (1 + r / 100)
    n_years = len(monthly_returns) / 12
    if n_years == 0:
        return 0.0
    return (product ** (1 / n_years) - 1) * 100

# ── Annualized volatility ────────────────────────────────────────────
def annual_vol(monthly_returns):
    if len(monthly_returns) < 2:
        return 0.0
    return np.std(monthly_returns, ddof=1) * math.sqrt(12)

# ── Sharpe ratio ─────────────────────────────────────────────────────
def sharpe(monthly_returns, rf=0.0):
    vol = annual_vol(monthly_returns)
    if vol == 0:
        return 0.0
    return (annualize(monthly_returns) - rf) / vol

# ── Main ─────────────────────────────────────────────────────────────
print("=" * 120)
print("NN SIGN BACKTEST — Full Statistical Rigor")
print("=" * 120)

for asset_name, (ticker, start_str) in ASSETS.items():
    print(f"\n{'#' * 120}")
    print(f"# {asset_name} ({ticker}) — {start_str} to 2026-07-01")
    print(f"{'#' * 120}")

    # ── Fetch price data ─────────────────────────────────────────────
    print(f"\nFetching {ticker}...")
    data = yf.download(ticker, start=start_str, end='2026-07-01', auto_adjust=True)
    closes = data['Close']
    if isinstance(closes, pd.DataFrame):
        closes = closes.iloc[:, 0]

    # Build price dict
    prices = {}
    for idx, val in closes.items():
        ts = int(idx.timestamp())
        day_start = ts - (ts % 86400)
        prices[day_start] = float(val)

    trading_dates = sorted([datetime.utcfromtimestamp(ts) for ts in prices.keys()])
    print(f"  {len(trading_dates)} trading days")

    # ── Compute NN positions ─────────────────────────────────────────
    start_dt = trading_dates[0] - timedelta(days=30)
    end_dt = trading_dates[-1] + timedelta(days=30)
    nn_positions = compute_nn_positions(start_dt, end_dt)
    sign_map = map_dates_to_sign(trading_dates, nn_positions)

    # ── Compute returns ──────────────────────────────────────────────
    # Daily returns
    daily_rets = {}
    for i in range(1, len(trading_dates)):
        d = trading_dates[i]
        d_str = d.strftime('%Y-%m-%d')
        prev = trading_dates[i-1]
        prev_str = prev.strftime('%Y-%m-%d')
        sign = sign_map.get(d_str)
        if sign and prev_str in sign_map:
            p_today = prices[int(d.timestamp()) - (int(d.timestamp()) % 86400)]
            p_yest = prices[int(prev.timestamp()) - (int(prev.timestamp()) % 86400)]
            if p_yest > 0:
                ret = (p_today - p_yest) / p_yest * 100
                daily_rets.setdefault(sign, []).append(ret)

    # Monthly returns
    monthly_rets = {}
    monthly_dates = {}
    for d in trading_dates:
        d_str = d.strftime('%Y-%m-%d')
        sign = sign_map.get(d_str)
        if sign:
            month_key = d.strftime('%Y-%m')
            if month_key not in monthly_dates:
                monthly_dates[month_key] = (d_str, sign)

    # Compute monthly returns from first to last trading day of each month
    month_prices = {}
    for d in trading_dates:
        d_str = d.strftime('%Y-%m-%d')
        month_key = d.strftime('%Y-%m')
        if month_key not in month_prices:
            month_prices[month_key] = {'first': (d_str, prices[int(d.timestamp()) - (int(d.timestamp()) % 86400)]),
                                        'last': (d_str, prices[int(d.timestamp()) - (int(d.timestamp()) % 86400)])}
        else:
            month_prices[month_key]['last'] = (d_str, prices[int(d.timestamp()) - (int(d.timestamp()) % 86400)])

    month_list = sorted(month_prices.keys())
    for i in range(1, len(month_list)):
        prev_m = month_list[i-1]
        curr_m = month_list[i]
        prev_last = month_prices[prev_m]['last'][1]
        curr_last = month_prices[curr_m]['last'][1]
        if prev_last > 0:
            ret = (curr_last - prev_last) / prev_last * 100
            sign = monthly_dates.get(curr_m, (None, None))[1]
            if sign:
                monthly_rets.setdefault(sign, []).append(ret)

    # ── Period-level returns ─────────────────────────────────────────
    # Find sign transitions
    transitions = []
    prev_sign = None
    for d_str in sorted(sign_map.keys()):
        sign = sign_map[d_str]
        if sign != prev_sign:
            transitions.append((d_str, sign))
            prev_sign = sign

    # Compute period returns
    period_rets = {}
    for i in range(len(transitions) - 1):
        start_str, sign = transitions[i]
        end_str, _ = transitions[i+1]
        start_d = datetime.strptime(start_str, '%Y-%m-%d')
        end_d = datetime.strptime(end_str, '%Y-%m-%d')

        # Find nearest trading days
        start_ts = int(start_d.timestamp()) - (int(start_d.timestamp()) % 86400)
        end_ts = int(end_d.timestamp()) - (int(end_d.timestamp()) % 86400)

        # Find actual trading day prices
        p_start = None
        p_end = None
        for ts in sorted(prices.keys()):
            if ts >= start_ts and p_start is None:
                p_start = prices[ts]
            if ts <= end_ts:
                p_end = prices[ts]

        if p_start and p_end and p_start > 0:
            ret = (p_end - p_start) / p_start * 100
            period_rets.setdefault(sign, []).append(ret)

    # ── BASELINE ─────────────────────────────────────────────────────
    all_daily = [r for rets in daily_rets.values() for r in rets]
    all_monthly = [r for rets in monthly_rets.values() for r in rets]
    all_period = [r for rets in period_rets.values() for r in rets]

    bl_daily_mean = np.mean(all_daily) if all_daily else 0
    bl_monthly_mean = np.mean(all_monthly) if all_monthly else 0
    bl_monthly_ann = annualize(all_monthly) if all_monthly else 0
    bl_monthly_vol = annual_vol(all_monthly) if all_monthly else 0
    bl_monthly_sharpe = sharpe(all_monthly) if all_monthly else 0

    print(f"\n  BASELINE (all months):")
    print(f"    Monthly mean: {bl_monthly_mean:+.3f}%")
    print(f"    Annualized:   {bl_monthly_ann:+.2f}%")
    print(f"    Ann Vol:      {bl_monthly_vol:.2f}%")
    print(f"    Sharpe:       {bl_monthly_sharpe:.2f}")
    print(f"    N months:     {len(all_monthly)}")

    # ── MONTHLY TABLE ────────────────────────────────────────────────
    print(f"\n  {'─' * 110}")
    print(f"  MONTHLY RETURNS BY NN SIGN")
    print(f"  {'─' * 110}")
    print(f"  {'Sign':<14} {'N':>5} {'Mean M':>8} {'Ann Ret':>8} {'Ann Vol':>8} {'Sharpe':>7} {'Win%':>6} "
          f"{'t-stat':>7} {'p-val':>7} {'d':>6} {'95% CI':>18}")
    print(f"  {'─' * 110}")

    monthly_results = []
    for sign in SIGNS:
        rets = monthly_rets.get(sign, [])
        if len(rets) < 3:
            continue
        mean_m = np.mean(rets)
        ann = annualize(rets)
        vol = annual_vol(rets)
        sh = sharpe(rets)
        win = sum(1 for r in rets if r > 0) / len(rets) * 100
        t_stat, p_val = stats.ttest_1samp(rets, bl_monthly_mean)
        d = cohens_d(rets, all_monthly)
        ci_lo, ci_hi = bootstrap_ci(rets)
        monthly_results.append((sign, len(rets), mean_m, ann, vol, sh, win, t_stat, p_val, d, ci_lo, ci_hi))
        print(f"  {sign:<14} {len(rets):>5} {mean_m:>+7.3f}% {ann:>+7.2f}% {vol:>7.2f}% {sh:>6.2f} {win:>5.0f}% "
              f"{t_stat:>+6.2f} {p_val:>7.4f} {d:>+5.2f} [{ci_lo:>+7.2f}%, {ci_hi:>+7.2f}%]")

    # ── PERIOD-LEVEL TABLE ───────────────────────────────────────────
    print(f"\n  {'─' * 100}")
    print(f"  PERIOD-LEVEL RETURNS (each ~1.5yr NN period = 1 observation)")
    print(f"  {'─' * 100}")
    print(f"  {'Sign':<14} {'N':>4} {'Mean':>8} {'Std':>8} {'Win%':>6} {'Min':>8} {'Max':>8} {'95% CI':>18}")
    print(f"  {'─' * 100}")

    period_results = []
    for sign in SIGNS:
        rets = period_rets.get(sign, [])
        if len(rets) < 2:
            continue
        mean_p = np.mean(rets)
        std_p = np.std(rets, ddof=1)
        win = sum(1 for r in rets if r > 0) / len(rets) * 100
        ci_lo, ci_hi = bootstrap_ci(rets)
        period_results.append((sign, len(rets), mean_p, std_p, win, min(rets), max(rets), ci_lo, ci_hi))
        print(f"  {sign:<14} {len(rets):>4} {mean_p:>+7.1f}% {std_p:>7.1f}% {win:>5.0f}% "
              f"{min(rets):>+7.1f}% {max(rets):>+7.1f}% [{ci_lo:>+7.1f}%, {ci_hi:>+7.1f}%]")

    # ── MANN-WHITNEY U: Best vs Worst ────────────────────────────────
    print(f"\n  {'─' * 100}")
    print(f"  MANN-WHITNEY U: Pairwise period-level comparisons")
    print(f"  {'─' * 100}")

    # Sort by period mean
    period_results.sort(key=lambda x: x[2], reverse=True)
    best_sign, best_n, best_mean = period_results[0][0], period_results[0][1], period_results[0][2]
    worst_sign, worst_n, worst_mean = period_results[-1][0], period_results[-1][1], period_results[-1][2]

    best_rets = period_rets.get(best_sign, [])
    worst_rets = period_rets.get(worst_sign, [])

    if len(best_rets) >= 3 and len(worst_rets) >= 3:
        u_stat, mw_p = stats.mannwhitneyu(best_rets, worst_rets, alternative='two-sided')
        d_bw = cohens_d(best_rets, worst_rets)
        print(f"  {best_sign} (mean={best_mean:+.1f}%, n={best_n}) vs {worst_sign} (mean={worst_mean:+.1f}%, n={worst_n})")
        print(f"    Mann-Whitney U = {u_stat:.1f}, p = {mw_p:.4f}, Cohen's d = {d_bw:+.2f}")

    # Also test best vs all others
    all_other_periods = [r for s, rets in period_rets.items() if s != best_sign for r in rets]
    if len(best_rets) >= 3 and len(all_other_periods) >= 3:
        u_stat2, mw_p2 = stats.mannwhitneyu(best_rets, all_other_periods, alternative='two-sided')
        d2 = cohens_d(best_rets, all_other_periods)
        print(f"  {best_sign} vs all other signs combined (n={len(all_other_periods)})")
        print(f"    Mann-Whitney U = {u_stat2:.1f}, p = {mw_p2:.4f}, Cohen's d = {d2:+.2f}")

    # ── FDR CORRECTION ───────────────────────────────────────────────
    print(f"\n  {'─' * 100}")
    print(f"  FDR CORRECTION (Benjamini-Hochberg) — Monthly t-tests")
    print(f"  {'─' * 100}")

    pvals = [(s, p, d) for s, n, _, _, _, _, _, _, p, d, _, _ in monthly_results]
    pvals.sort(key=lambda x: x[1])
    n_tests = len(pvals)
    fdr_thresholds = [(i + 1) / n_tests * 0.05 for i in range(n_tests)]

    print(f"  {'Sign':<14} {'p-val':>8} {'Rank':>5} {'FDR thresh':>12} {'Survives?':>10} {'d':>6}")
    print(f"  {'─' * 60}")
    survivors = []
    for rank, (sign, p, d) in enumerate(pvals):
        thresh = fdr_thresholds[rank]
        survives = p < thresh
        if survives:
            survivors.append((sign, p, d))
        print(f"  {sign:<14} {p:>8.4f} {rank+1:>5} {thresh:>12.6f} {'YES' if survives else 'no':>10} {d:>+5.2f}")

    if survivors:
        print(f"\n  FDR survivors: {', '.join(f'{s} (p={p:.4f}, d={d:+.2f})' for s, p, d in survivors)}")
    else:
        print(f"\n  No signs survive FDR correction.")

    # ── OUT-OF-SAMPLE ────────────────────────────────────────────────
    print(f"\n  {'─' * 100}")
    print(f"  OUT-OF-SAMPLE: Train pre-{OOS_SPLIT.strftime('%Y')}, test post-{OOS_SPLIT.strftime('%Y')}")
    print(f"  {'─' * 100}")

    oos_monthly = {}
    for sign in SIGNS:
        rets = monthly_rets.get(sign, [])
        # We need dates for each monthly return — rebuild with dates
        pass

    # Rebuild monthly with dates
    monthly_with_dates = {}
    month_list_sorted = sorted(month_prices.keys())
    for i in range(1, len(month_list_sorted)):
        prev_m = month_list_sorted[i-1]
        curr_m = month_list_sorted[i]
        prev_last = month_prices[prev_m]['last'][1]
        curr_last = month_prices[curr_m]['last'][1]
        if prev_last > 0:
            ret = (curr_last - prev_last) / prev_last * 100
            sign = monthly_dates.get(curr_m, (None, None))[1]
            if sign:
                curr_dt = datetime.strptime(curr_m + '-01', '%Y-%m-%d')
                monthly_with_dates.setdefault(sign, []).append((curr_dt, ret))

    # Split
    train_monthly = {s: [r for d, r in v if d < OOS_SPLIT] for s, v in monthly_with_dates.items()}
    test_monthly = {s: [r for d, r in v if d >= OOS_SPLIT] for s, v in monthly_with_dates.items()}

    # Train: find best and worst signs
    train_means = {}
    for sign in SIGNS:
        rets = train_monthly.get(sign, [])
        if rets:
            train_means[sign] = np.mean(rets)

    if train_means:
        best_train = max(train_means, key=train_means.get)
        worst_train = min(train_means, key=train_means.get)

        # Test: how did they do?
        test_best = test_monthly.get(best_train, [])
        test_worst = test_monthly.get(worst_train, [])

        print(f"  Train best:  {best_train} (mean={train_means[best_train]:+.3f}%, n={len(train_monthly[best_train])})")
        print(f"  Train worst: {worst_train} (mean={train_means[worst_train]:+.3f}%, n={len(train_monthly[worst_train])})")

        if test_best:
            test_mean_best = np.mean(test_best)
            test_ann_best = annualize(test_best)
            t_best, p_best = stats.ttest_1samp(test_best, bl_monthly_mean)
            print(f"  Test {best_train}: mean={test_mean_best:+.3f}%, ann={test_ann_best:+.2f}%, "
                  f"n={len(test_best)}, p={p_best:.4f} vs baseline")

        if test_worst:
            test_mean_worst = np.mean(test_worst)
            test_ann_worst = annualize(test_worst)
            t_worst, p_worst = stats.ttest_1samp(test_worst, bl_monthly_mean)
            print(f"  Test {worst_train}: mean={test_mean_worst:+.3f}%, ann={test_ann_worst:+.2f}%, "
                  f"n={len(test_worst)}, p={p_worst:.4f} vs baseline")

        # OOS spread
        if test_best and test_worst:
            spread = test_mean_best - test_mean_worst
            # Bootstrap the spread
            spreads = []
            for _ in range(BOOTSTRAP_N):
                b_best = np.random.choice(test_best, size=len(test_best), replace=True)
                b_worst = np.random.choice(test_worst, size=len(test_worst), replace=True)
                spreads.append(np.mean(b_best) - np.mean(b_worst))
            spread_lo = np.percentile(spreads, 2.5)
            spread_hi = np.percentile(spreads, 97.5)
            print(f"  OOS spread ({best_train} - {worst_train}): {spread:+.3f}% monthly "
                  f"[{spread_lo:+.3f}%, {spread_hi:+.3f}%]")

    # ── NN TIMELINE ──────────────────────────────────────────────────
    print(f"\n  {'─' * 100}")
    print(f"  NN SIGN TIMELINE (recent + upcoming)")
    print(f"  {'─' * 100}")

    # Compute from 2020 onward
    timeline_start = datetime(2020, 1, 1)
    timeline_end = datetime(2035, 1, 1)
    tl_positions = compute_nn_positions(timeline_start, timeline_end)
    tl_transitions = []
    prev = None
    for d_str in sorted(tl_positions.keys()):
        s = tl_positions[d_str]
        if s != prev:
            tl_transitions.append((d_str, s))
            prev = s

    for i in range(len(tl_transitions)):
        d_str, sign = tl_transitions[i]
        next_str = tl_transitions[i+1][0] if i+1 < len(tl_transitions) else '?'
        print(f"  {sign:<14} {d_str} → {next_str}")

    # ── SUMMARY ──────────────────────────────────────────────────────
    print(f"\n  {'─' * 100}")
    print(f"  SUMMARY: {asset_name}")
    print(f"  {'─' * 100}")

    # Rank by annualized return
    monthly_results.sort(key=lambda x: x[3], reverse=True)
    print(f"  Ranked by annualized return:")
    for sign, n, mean_m, ann, vol, sh, win, t, p, d, ci_lo, ci_hi in monthly_results:
        sig = "***" if p < 0.001 else "**" if p < 0.01 else "*" if p < 0.05 else ""
        print(f"    {sign:<14} {ann:>+7.2f}% ann  Sharpe={sh:>5.2f}  n={n:>4}  p={p:.4f}{sig}  d={d:>+5.2f}  "
              f"CI=[{ci_lo:>+6.2f}%, {ci_hi:>+6.2f}%]")

print("\n" + "=" * 120)
print("DONE")
print("=" * 120)
