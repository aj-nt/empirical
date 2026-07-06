#!/usr/bin/env python3
"""
Uranus-Saturn Transit Backtest — OOS Validation (Batch-Optimized)

Tests: Uranus square/trine natal Saturn → abnormal positive returns (5d horizon).

Original signal (prior backtest, ~500 stocks, 26 years):
  - Uranus square natal Saturn, 5d: +3.60% abnormal, d=+0.352, FDR p=0.028, n=122
  - Uranus trine natal Saturn, 5d: +2.83% abnormal, d=+0.352, FDR p=0.004, n=164

This version batch-downloads all price data upfront to avoid per-event yfinance calls.
"""

import swisseph as swe
import yfinance as yf
import pandas as pd
import numpy as np
from scipy import stats
from datetime import datetime, timedelta
from collections import defaultdict
import warnings
warnings.filterwarnings('ignore')

# ─── Configuration ───────────────────────────────────────────────────────────

ASPECTS = {
    'conjunction': 0,
    'trine': 120,
    'square': 90,
    'opposition': 180,
}
ORB = 2.0
HORIZON_DAYS = 5
TRAIN_CUTOFF = '2020-01-01'
MIN_HISTORY_YEARS = 5
BATCH_SIZE = 100  # tickers per yfinance batch download

# ─── Helpers ─────────────────────────────────────────────────────────────────

def get_sp500_tickers():
    import requests, io
    headers = {'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'}
    resp = requests.get('https://en.wikipedia.org/wiki/List_of_S%26P_500_companies', headers=headers, timeout=30)
    table = pd.read_html(io.StringIO(resp.text))[0]
    return table['Symbol'].tolist()


def compute_natal_saturn(date):
    jd = swe.julday(date.year, date.month, date.day, 12.0)
    lon, _ = swe.calc_ut(jd, swe.SATURN)
    return lon[0]


def compute_uranus_positions(start_date, end_date):
    dates = pd.date_range(start_date, end_date, freq='D')
    positions = []
    for d in dates:
        jd = swe.julday(d.year, d.month, d.day, 12.0)
        lon, _ = swe.calc_ut(jd, swe.URANUS)
        positions.append(lon[0])
    return pd.Series(positions, index=dates)


def is_aspect(uranus_lon, natal_saturn_lon, aspect_name, orb=ORB):
    target = ASPECTS[aspect_name]
    diff = abs((uranus_lon - natal_saturn_lon) % 360)
    for base in [target, 360 - target]:
        if abs(diff - base) <= orb:
            return True
    return False


def find_aspect_dates(uranus_series, natal_saturn_lon, aspect_name, orb=ORB):
    hits = []
    for date, uranus_lon in uranus_series.items():
        if is_aspect(uranus_lon, natal_saturn_lon, aspect_name, orb):
            hits.append(date)
    return hits


def bootstrap_ci(data, n_resamples=10000, ci=95):
    if len(data) < 2:
        return np.nan, np.nan
    rng = np.random.default_rng(42)
    means = []
    for _ in range(n_resamples):
        sample = rng.choice(data, size=len(data), replace=True)
        means.append(np.mean(sample))
    low = (100 - ci) / 2
    high = 100 - low
    return np.percentile(means, low), np.percentile(means, high)


def cohens_d(x, y):
    nx, ny = len(x), len(y)
    if nx < 2 or ny < 2:
        return np.nan
    pooled_std = np.sqrt(((nx - 1) * np.var(x, ddof=1) + (ny - 1) * np.var(y, ddof=1)) / (nx + ny - 2))
    if pooled_std == 0:
        return np.nan
    return (np.mean(x) - np.mean(y)) / pooled_std


def temporal_persistence(data, n_splits=2):
    if len(data) < 4:
        return False
    split = len(data) // n_splits
    signs = []
    for i in range(n_splits):
        chunk = data[i * split:(i + 1) * split]
        if len(chunk) > 0:
            signs.append(np.mean(chunk) > 0)
    return all(s == signs[0] for s in signs)


# ─── Main ────────────────────────────────────────────────────────────────────

def main():
    print("=" * 80)
    print("URANUS-SATURN TRANSIT BACKTEST — OOS VALIDATION")
    print("=" * 80)
    print(f"\nAspects: conjunction, trine, square, opposition")
    print(f"Orb: {ORB}° | Horizon: {HORIZON_DAYS}d | Split: pre-{TRAIN_CUTOFF} / {TRAIN_CUTOFF}+")
    print()

    # 1. Get S&P 500 tickers
    print("Fetching S&P 500 constituents...")
    tickers = get_sp500_tickers()
    print(f"  {len(tickers)} tickers")

    # 2. Batch-download all price data
    print("\nBatch-downloading price data (this will take a while)...")
    all_prices = {}
    for i in range(0, len(tickers), BATCH_SIZE):
        batch = tickers[i:i + BATCH_SIZE]
        print(f"  Batch {i // BATCH_SIZE + 1}: {len(batch)} tickers...", end=' ', flush=True)
        try:
            data = yf.download(batch, start='1990-01-01', auto_adjust=True, progress=False, threads=True)
            for t in batch:
                col = ('Close', t)
                if col in data.columns:
                    series = data[col].dropna()
                    if len(series) > 0:
                        all_prices[t] = series
            print(f"got {sum(1 for t in batch if t in all_prices)}")
        except Exception as e:
            print(f"error: {e}")
    print(f"  Total: {len(all_prices)} tickers with price data")

    # 3. Get SPY for abnormal returns
    print("\nFetching SPY...")
    spy = yf.download('SPY', start='1990-01-01', auto_adjust=True, progress=False)
    spy_close = spy[('Close', 'SPY')]
    spy_cumret = spy_close.pct_change().cumsum()
    print(f"  SPY: {len(spy_close)} days")

    # 4. Compute Uranus positions
    print("\nComputing Uranus positions (1990–2026)...")
    uranus = compute_uranus_positions(datetime(1990, 1, 1), datetime(2026, 7, 2))
    print(f"  {len(uranus)} daily positions")

    # 5. For each stock: get inception from price data, compute natal Saturn, find hits
    print("\nProcessing events...")
    results = {aspect: {'train': [], 'test': [], 'all': []} for aspect in ASPECTS}
    stock_count = 0
    event_count = 0

    for ticker, prices in all_prices.items():
        inception = prices.index[0].to_pydatetime()
        years_of_data = 2026 - inception.year
        if years_of_data < MIN_HISTORY_YEARS:
            continue

        natal_saturn = compute_natal_saturn(inception)

        for aspect_name in ASPECTS:
            hits = find_aspect_dates(uranus, natal_saturn, aspect_name, ORB)
            for hit_date in hits:
                if hit_date < inception + timedelta(days=30):
                    continue
                if hit_date > datetime(2026, 6, 25):
                    continue

                # Find closest trading day >= hit_date
                future_prices = prices[prices.index >= hit_date]
                if len(future_prices) < 2:
                    continue
                start_price = future_prices.iloc[0]

                # Find price HORIZON_DAYS later
                target = hit_date + timedelta(days=HORIZON_DAYS)
                future_prices2 = prices[prices.index >= target]
                if len(future_prices2) == 0:
                    continue
                end_price = future_prices2.iloc[0]

                stock_ret = (end_price - start_price) / start_price

                # SPY return over same period
                spy_future = spy_cumret[spy_cumret.index >= hit_date]
                spy_target = spy_cumret[spy_cumret.index >= target]
                if len(spy_future) == 0 or len(spy_target) == 0:
                    continue
                spy_ret = spy_target.iloc[0] - spy_future.iloc[0]

                ab_ret = stock_ret - spy_ret
                is_train = hit_date < pd.Timestamp(TRAIN_CUTOFF)
                bucket = 'train' if is_train else 'test'
                results[aspect_name][bucket].append(ab_ret)
                results[aspect_name]['all'].append(ab_ret)
                event_count += 1

        stock_count += 1
        if stock_count % 100 == 0:
            print(f"  {stock_count} stocks, {event_count} events...")

    print(f"\n  Done. {stock_count} stocks, {event_count} total events")

    # 6. Report
    print("\n" + "=" * 80)
    print("RESULTS")
    print("=" * 80)

    for aspect_name in ['square', 'trine', 'conjunction', 'opposition']:
        r = results[aspect_name]
        train_data = np.array(r['train'])
        test_data = np.array(r['test'])
        all_data = np.array(r['all'])

        print(f"\n─── Uranus {aspect_name} natal Saturn ───")
        print(f"  Train (pre-{TRAIN_CUTOFF}): n={len(train_data)}")
        print(f"  Test  ({TRAIN_CUTOFF}+):    n={len(test_data)}")

        if len(train_data) >= 5:
            train_mean = np.mean(train_data) * 100
            train_ci = bootstrap_ci(train_data)
            train_t, train_p = stats.ttest_1samp(train_data, 0)
            print(f"  Train mean: {train_mean:+.2f}%  [{train_ci[0]*100:+.2f}%, {train_ci[1]*100:+.2f}%]")
            print(f"  Train t-test vs 0: t={train_t:.3f}, p={train_p:.4f}")
            print(f"  Train % positive: {np.mean(train_data > 0)*100:.1f}%")
            print(f"  Train temporal persistence: {temporal_persistence(train_data)}")

        if len(test_data) >= 5:
            test_mean = np.mean(test_data) * 100
            test_ci = bootstrap_ci(test_data)
            test_t, test_p = stats.ttest_1samp(test_data, 0)
            print(f"  Test mean:  {test_mean:+.2f}%  [{test_ci[0]*100:+.2f}%, {test_ci[1]*100:+.2f}%]")
            print(f"  Test t-test vs 0: t={test_t:.3f}, p={test_p:.4f}")
            print(f"  Test % positive: {np.mean(test_data > 0)*100:.1f}%")

        if len(train_data) >= 5 and len(test_data) >= 5:
            d = cohens_d(train_data, test_data)
            mw_u, mw_p = stats.mannwhitneyu(train_data, test_data, alternative='two-sided')
            print(f"  Train vs Test Cohen's d: {d:+.3f}")
            print(f"  Train vs Test MWU p: {mw_p:.4f}")

        if len(all_data) >= 5:
            all_mean = np.mean(all_data) * 100
            all_ci = bootstrap_ci(all_data)
            print(f"  All: {all_mean:+.2f}% [{all_ci[0]*100:+.2f}%, {all_ci[1]*100:+.2f}%]  n={len(all_data)}  pos={np.mean(all_data > 0)*100:.1f}%")

    # 7. Verdict
    print("\n" + "=" * 80)
    print("VERDICT")
    print("=" * 80)

    for aspect_name in ['square', 'trine']:
        r = results[aspect_name]
        train_data = np.array(r['train'])
        test_data = np.array(r['test'])

        if len(test_data) < 5:
            print(f"\nUranus {aspect_name} Saturn: INSUFFICIENT TEST DATA (n={len(test_data)})")
            continue

        test_mean = np.mean(test_data) * 100
        test_t, test_p = stats.ttest_1samp(test_data, 0)
        test_pos = np.mean(test_data > 0) * 100

        if test_p < 0.05 and test_mean > 0:
            print(f"\nUranus {aspect_name} Saturn: SURVIVES OOS — {test_mean:+.2f}%, p={test_p:.4f}, {test_pos:.0f}% pos, n={len(test_data)}")
        elif test_mean > 0:
            print(f"\nUranus {aspect_name} Saturn: DIRECTION CORRECT, not significant — {test_mean:+.2f}%, p={test_p:.4f}, {test_pos:.0f}% pos, n={len(test_data)}")
        else:
            print(f"\nUranus {aspect_name} Saturn: DEAD — {test_mean:+.2f}%, p={test_p:.4f}, {test_pos:.0f}% pos, n={len(test_data)}")

    print("\nDone.")


if __name__ == '__main__':
    main()
