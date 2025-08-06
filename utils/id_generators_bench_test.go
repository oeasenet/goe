package utils

import (
	"testing"
)

// Benchmark for NanoID generation with default length
func BenchmarkGenerateNanoId_Default(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateNanoId()
	}
}

// Benchmark for NanoID generation with different lengths
func BenchmarkGenerateNanoId_Length10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateNanoId(10)
	}
}

func BenchmarkGenerateNanoId_Length21(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateNanoId(21)
	}
}

func BenchmarkGenerateNanoId_Length50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateNanoId(50)
	}
}

func BenchmarkGenerateNanoId_Length100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateNanoId(100)
	}
}

// Memory allocation benchmarks for NanoID
func BenchmarkGenerateNanoId_Allocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		GenerateNanoId()
	}
}

func BenchmarkGenerateNanoId_AllocsLength50(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		GenerateNanoId(50)
	}
}

// Parallel benchmarks to test concurrent performance for NanoID
func BenchmarkGenerateNanoId_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateNanoId()
		}
	})
}

func BenchmarkGenerateNanoId_ParallelLength21(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateNanoId(21)
		}
	})
}

// Batch generation benchmarks for NanoID
func BenchmarkGenerateNanoId_Batch100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			GenerateNanoId()
		}
	}
}

func BenchmarkGenerateNanoId_Batch1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			GenerateNanoId()
		}
	}
}

// Stress test benchmarks - high volume generation for NanoID
func BenchmarkGenerateNanoId_Stress(b *testing.B) {
	b.SetBytes(32) // Default nanoid length
	for i := 0; i < b.N; i++ {
		GenerateNanoId()
	}
}

// Comprehensive comparison benchmark - all three generators in one test
func BenchmarkIDGenerators_Comparison(b *testing.B) {
	b.Run("NanoID_Default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateNanoId()
		}
	})

	b.Run("NanoID_Length21", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateNanoId(21)
		}
	})

	b.Run("UUIDv7", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateUUIDv7()
		}
	})

	b.Run("XID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateXid()
		}
	})
}

// Parallel comparison benchmark
func BenchmarkIDGenerators_ParallelComparison(b *testing.B) {
	b.Run("NanoID_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				GenerateNanoId()
			}
		})
	})

	b.Run("UUIDv7_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				GenerateUUIDv7()
			}
		})
	})

	b.Run("XID_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				GenerateXid()
			}
		})
	})
}

// Memory allocation comparison benchmark
func BenchmarkIDGenerators_AllocComparison(b *testing.B) {
	b.Run("NanoID_Allocs", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenerateNanoId()
		}
	})

	b.Run("UUIDv7_Allocs", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenerateUUIDv7()
		}
	})

	b.Run("XID_Allocs", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenerateXid()
		}
	})
}
