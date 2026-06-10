# Supplementary Material: Uranian Transit Backtest

This analysis uses the Hamburg School Uranian system, a 20th-century framework distinct from the Hellenistic-based measurements in the main paper. The system employs hypothetical trans-Neptunian points (Vulkanus, Hades, Kronos, Admetos, Cupido, Poseidon), not classical planets. This material is included for completeness as the study's only external validation result.

## Data and Method

SPY (S&P 500 ETF) daily OHLC data from April 2021 to April 2026. Two resolutions tested.

### Daily Resolution

1,262 trading days. Continuous point score computed from exact transit hits, normalized per day. Score showed in-sample correlation with daily range (r = 0.34), but the signal inverted out of sample. Points predicting high volatility in 2021-2023 predicted low volatility in 2024-2026. Score autocorrelation of 0.93 at lag 1 indicated slow-moving regime measurement, not daily event detection.

Daily resolution produced no usable signal.

### Weekly Resolution

262 weeks. Discrete net signal computed from exact transit hits. Point classifications were a priori from tradition: high-volatility (Vulkanus, Hades, Kronos) and low-volatility (Admetos, Cupido, Poseidon).

Results: HV weeks (net signal positive, n = 60) showed 35% high-volatility rate versus 7% for LV weeks (net signal negative, n = 46). Quarterly direction was correct in 10 of 12 quarters (83%, p = 0.017 by permutation test). Pairwise comparison within quarters: 68.5% of HV-LV pairs had HV with higher range (p = 0.009).

Year 2025 (April tariff shock during an LV week) weakened but did not eliminate the effect.

## Interpretation

The contrast between daily and weekly resolutions suggests the Uranian transit signal operates at a weekly timescale. The daily score was confounded by autocorrelation. The weekly net signal, using discrete exact hits rather than continuous ratios, revealed a statistically significant but modest relationship between transit point activations and market volatility.

This is a single-domain result with a single methodology. Replication across indices, time periods, and point sets is required before the signal can be considered robust.

## Reference

Witte, A. and Lefeldt, H. (1928). Rules for Planetary Pictures. Hamburg School.
