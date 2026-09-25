-- Verinode Seed Migration: 000002_seed_initial_grade.up.sql
-- Seeds the initial standardized benchmark delivery grade: 8x H100 SXM 80GB, 168h

INSERT INTO grades (
    id,
    gpu_sku,
    min_memory_gb,
    topology,
    min_healthy_gpu_count,
    benchmark_floor,
    min_cpu_cores,
    min_ram_gb,
    min_nvme_perf
) VALUES (
    'H100-SXM-8XNV',
    'NVIDIA H100 SXM 80GB',
    640,
    'SXM5/HGX 8x NVLink 4.0 / NVSwitch (900 GB/s bidirectional)',
    8,
    '{"nccl_allreduce_gb_per_sec": 400.0, "min_cuda_driver": "535.129.03", "max_ecc_unrecovered_errors": 0}',
    112,
    1024,
    100000
) ON CONFLICT (id) DO NOTHING;
