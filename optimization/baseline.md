# ffire Benchmark Comparison

Generated: 2025-11-05T20:59:23+01:00

| Language | Format | Message | Encode (ns) | Decode (ns) | Total (ns) | Wire Size |
|----------|--------|---------|-------------|-------------|------------|----------|
| Go | ffire | array_float | 1382 | 5279 | 6661 | 20002 |
| Go | protobuf | array_float | 3617 | 10928 | 14545 | 20004 |
| Go | ffire | array_int | 1495 | 5359 | 6854 | 20002 |
| Go | protobuf | array_int | 18307 | 12104 | 30411 | 9876 |
| Go | ffire | array_string | 10280 | 7973 | 18253 | 17002 |
| Go | protobuf | array_string | 5482 | 13115 | 18597 | 17000 |
| Go | ffire | array_struct | 7089 | 3294 | 10383 | 5202 |
| Go | protobuf | array_struct | 7899 | 16033 | 23932 | 5200 |
| Go | ffire | complex | 7593 | 4698 | 12291 | 4293 |
| Go | protobuf | complex | 8158 | 13115 | 21273 | 3901 |
| Go | ffire | empty | 35 | 7 | 42 | 8 |
| Go | protobuf | empty | 43 | 67 | 110 | 0 |
| Go | ffire | nested | 1525 | 5455 | 6980 | 20002 |
| Go | protobuf | nested | 18548 | 13378 | 31926 | 9902 |
| Go | ffire | optional | 24803 | 35323 | 60126 | 21840 |
| Go | protobuf | optional | 44173 | 97117 | 141290 | 20996 |
| Go | ffire | struct | 52 | 19 | 71 | 24 |
| Go | protobuf | struct | 102 | 105 | 207 | 23 |
| Go | ffire | tags | 45 | 28 | 73 | 34 |
| Go | protobuf | tags | 85 | 120 | 205 | 28 |

## Notes

- All benchmarks use the same test fixture
- Measurements exclude warmup and fixture loading
- Results are averaged over multiple iterations
