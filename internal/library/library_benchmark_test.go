package library

import (
	"testing"
)

func BenchmarkResolveBuiltin(b *testing.B) {
	manager := NewManager("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.Resolve("patterns.rest_api")
		if err != nil {
			b.Fatalf("Resolve failed: %v", err)
		}
	}
}

func BenchmarkListLibraries(b *testing.B) {
	manager := NewManager("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ListLibraries()
	}
}

func BenchmarkValidateImportName(b *testing.B) {
	testNames := []string{
		"patterns.rest_api",
		"my-library.sub_module.test",
		"simple",
		"very.long.nested.import.path.to.test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testNames {
			_ = validateImportName(name)
		}
	}
}

func BenchmarkLoadEmbedded(b *testing.B) {
	manager := NewManager("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.loadEmbedded("patterns/rest_api.prompt")
		if err != nil {
			b.Fatalf("loadEmbedded failed: %v", err)
		}
	}
}
