#!/usr/bin/env python3
"""
Mundane Astrology Backtesting Harness

Tests astrological claims against market data.
Loads mundane events from JSON export, fetches market data via yfinance,
computes returns around events, and compares against random baselines.

Usage:
    python3 backtest_mundane.py --events /tmp/mundane_events.json
"""

import json
import sys
import argparse
import random
import statistics
from datetime import datetime, timedelta
from collections import defaultdict

import yfinance as yf
import numpy as np


def load_events(path):
    with open(path) as f:
        return json.load(f)


def fetch_market_data(symbol, start, end):
    """Fetch daily adjusted close prices."""
    ticker = yf.Ticker(symbol)
    df = ticker.history(start=start, end=end, auto_adjust=True)
    if df.empty:
        print(f"WARNING: No data for {symbol} {start} to {end}")
        return None
    return df


def precompute_returns(df, window_days):
    """
    Pre-compute forward returns for every trading day.
    Returns dict: date -> return_pct
    """
    import pandas as pd
    dates = df.index.tz_localize(None) if df.index.tz is not None else df.index
    closes = df['Close'].values
    
    returns = {}
    for i in range(len(dates)):
        # Find the price window_days later
        target_date = dates[i] + timedelta(days=window_days)
        # Find closest trading day to target
        for j in range(i + 1, len(dates)):
            if dates[j] >= target_date:
                ret = (closes[j] / closes[i] - 1) * 100
                returns[dates[i]] = ret
                break
    
    return returns


def compute_returns_fast(precomputed, date):
    """
    Look up forward return from pre-computed dict.
    Returns (return_pct, success).
    """
    import pandas as pd
    # Find closest trading day
    for i in range(-5, 6):
        d = pd.Timestamp(date + timedelta(days=i))
        if d in precomputed:
            return precomputed[d], True
    return None, False


def random_baseline_fast(precomputed, n_samples, n_trials=1000):
    """
    Fast random baseline using pre-computed returns.
    """
    all_returns = list(precomputed.values())
    if len(all_returns) < n_samples:
        return None
    
    trial_means = []
    for _ in range(n_trials):
        sample = random.sample(all_returns, n_samples)
        trial_means.append(statistics.mean(sample))
    
    return {
        'mean': statistics.mean(trial_means),
        'std': statistics.stdev(trial_means) if len(trial_means) > 1 else 0,
        'pctile_5': np.percentile(trial_means, 5),
        'pctile_95': np.percentile(trial_means, 95),
    }


def backtest_events_fast(events, precomputed, event_name):
    """
    Fast backtest using pre-computed returns.
    """
    returns = []
    for event in events:
        date_str = event['date']
        date = datetime.strptime(date_str, '%Y-%m-%d')
        ret, ok = compute_returns_fast(precomputed, date)
        if ok:
            returns.append(ret)
    
    if not returns:
        return None
    
    return {
        'event': event_name,
        'count': len(returns),
        'mean_return': statistics.mean(returns),
        'median_return': statistics.median(returns),
        'std_return': statistics.stdev(returns) if len(returns) > 1 else 0,
        'min_return': min(returns),
        'max_return': max(returns),
        'positive_pct': sum(1 for r in returns if r > 0) / len(returns) * 100,
        'returns': returns,
    }


def ttest_against_baseline(event_returns, baseline_mean, baseline_std, n_events):
    """Simple z-test: is the event mean significantly different from baseline?"""
    if baseline_std == 0 or n_events < 2:
        return None
    
    event_mean = statistics.mean(event_returns)
    event_std = statistics.stdev(event_returns)
    event_se = event_std / np.sqrt(n_events)
    
    # Z-score: how many standard errors is the event mean from baseline?
    z = (event_mean - baseline_mean) / event_se
    
    # Two-tailed p-value approximation
    from math import erf, sqrt
    def norm_cdf(x):
        return 0.5 * (1 + erf(x / sqrt(2)))
    
    p_value = 2 * (1 - norm_cdf(abs(z)))
    
    return {
        'z_score': z,
        'p_value': p_value,
        'significant_05': p_value < 0.05,
        'significant_01': p_value < 0.01,
    }


def backtest_conjunctions_by_sign(events, df, planet1, planet2, window_days=30):
    """Backtest conjunctions, grouping by the sign they occur in."""
    # We need to compute the sign. For conjunctions, we can use the
    # longitude from the event data. But our export doesn't include lon.
    # Instead, group by date and compute sign from SWE.
    # For now, just do the overall backtest.
    return backtest_events(events, df, f"{planet1}-{planet2} conjunctions", window_days)


def main():
    parser = argparse.ArgumentParser(description='Mundane astrology backtesting')
    parser.add_argument('--events', required=True, help='JSON events file')
    parser.add_argument('--symbol', default='^GSPC', help='Market symbol (default: ^GSPC)')
    parser.add_argument('--window', type=int, default=30, help='Forward return window in days')
    parser.add_argument('--trials', type=int, default=1000, help='Random baseline trials')
    args = parser.parse_args()
    
    print(f"Loading events from {args.events}...")
    data = load_events(args.events)
    
    start = data['start_date']
    end = data['end_date']
    print(f"Event range: {start} to {end}")
    print(f"Events: {len(data['solar_ingresses'])} ingresses, "
          f"{len(data['lunations'])} lunations, "
          f"{len(data['eclipses'])} eclipses, "
          f"{len(data['conjunctions'])} conjunctions, "
          f"{len(data['planetary_ingresses'])} planetary ingresses")
    
    # Fetch market data
    print(f"\nFetching {args.symbol} data...")
    df = fetch_market_data(args.symbol, start, end)
    if df is None:
        print("ERROR: Could not fetch market data")
        sys.exit(1)
    print(f"Got {len(df)} trading days from {df.index[0].date()} to {df.index[-1].date()}")
    
    # Pre-compute all forward returns
    print(f"\nPre-computing {args.window}-day forward returns...")
    precomputed = precompute_returns(df, args.window)
    print(f"Pre-computed {len(precomputed)} returns")
    
    results = []
    
    def run_test(events, name):
        """Run a single backtest with baseline comparison."""
        if not events:
            return
        result = backtest_events_fast(events, precomputed, name)
        if not result:
            return
        baseline = random_baseline_fast(precomputed, result['count'], args.trials)
        if not baseline:
            return
        test = ttest_against_baseline(
            result['returns'], baseline['mean'], baseline['std'], result['count'])
        
        print(f"\n  {name}:")
        print(f"    Events: {result['count']}")
        print(f"    Mean {args.window}d return: {result['mean_return']:.2f}%")
        print(f"    Median: {result['median_return']:.2f}%")
        print(f"    Std: {result['std_return']:.2f}%")
        print(f"    Positive: {result['positive_pct']:.1f}%")
        print(f"    Range: {result['min_return']:.2f}% to {result['max_return']:.2f}%")
        print(f"    Random baseline: {baseline['mean']:.2f}% (±{baseline['std']:.2f}%)")
        print(f"    Baseline 5th-95th: {baseline['pctile_5']:.2f}% to {baseline['pctile_95']:.2f}%")
        if test:
            sig = "***" if test['significant_01'] else ("**" if test['significant_05'] else "")
            print(f"    Z-score: {test['z_score']:.2f}  P-value: {test['p_value']:.4f} {sig}")
        
        result['baseline'] = baseline
        result['test'] = test
        results.append(result)
    
    # ── Test 1: Saturn-Jupiter conjunctions ──
    print("\n" + "="*60)
    print("TEST 1: Saturn-Jupiter Conjunctions vs Market Returns")
    print("="*60)
    
    sj_conj = [e for e in data['conjunctions'] 
               if (e['planet1'] == 'Jupiter' and e['planet2'] == 'Saturn') or
                  (e['planet1'] == 'Saturn' and e['planet2'] == 'Jupiter')]
    print(f"Found {len(sj_conj)} Saturn-Jupiter conjunctions")
    run_test(sj_conj, "Saturn-Jupiter")
    
    # ── Test 2: Eclipses vs Market Returns ──
    print("\n" + "="*60)
    print("TEST 2: Eclipses vs Market Returns")
    print("="*60)
    
    for etype in ['Solar Eclipse', 'Lunar Eclipse']:
        ecl = [e for e in data['eclipses'] if e['type'] == etype]
        print(f"\n{etype}s: {len(ecl)} events")
        run_test(ecl, etype)
    
    # ── Test 3: Mars-Saturn conjunctions ──
    print("\n" + "="*60)
    print("TEST 3: Mars-Saturn Conjunctions vs Market Returns")
    print("="*60)
    
    ms_conj = [e for e in data['conjunctions']
               if (e['planet1'] == 'Mars' and e['planet2'] == 'Saturn') or
                  (e['planet1'] == 'Saturn' and e['planet2'] == 'Mars')]
    print(f"Found {len(ms_conj)} Mars-Saturn conjunctions")
    run_test(ms_conj, "Mars-Saturn")
    
    # ── Test 4: Solar Ingresses ──
    print("\n" + "="*60)
    print("TEST 4: Solar Ingresses (Equinoxes/Solstices) vs Market Returns")
    print("="*60)
    
    for sign in ['Aries', 'Cancer', 'Libra', 'Capricorn']:
        ing = [e for e in data['solar_ingresses'] if e['sign'] == sign]
        run_test(ing, f"Sun→{sign}")
    
    # ── Test 5: New Moon vs Full Moon ──
    print("\n" + "="*60)
    print("TEST 5: New Moon vs Full Moon Returns")
    print("="*60)
    
    for ltype in ['New Moon', 'Full Moon']:
        lun = [e for e in data['lunations'] if e['type'] == ltype]
        run_test(lun, ltype)
    
    # ── Summary ──
    print("\n" + "="*60)
    print("SUMMARY")
    print("="*60)
    
    significant = [r for r in results if r.get('test') and r['test']['significant_05']]
    if significant:
        print(f"\nStatistically significant results (p<0.05):")
        for r in significant:
            sig_level = "p<0.01" if r['test']['significant_01'] else "p<0.05"
            print(f"  {r['event']}: mean={r['mean_return']:.2f}%, "
                  f"p={r['test']['p_value']:.4f} ({sig_level}), "
                  f"n={r['count']}")
    else:
        print("\nNo statistically significant results found.")
    
    print(f"\nTotal tests run: {len(results)}")
    print(f"Market: {args.symbol}")
    print(f"Window: {args.window} days")
    print(f"Baseline trials: {args.trials}")


if __name__ == '__main__':
    import pandas as pd
    main()
