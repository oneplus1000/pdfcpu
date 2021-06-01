package test

import (
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func TestDrawLineFile(t *testing.T) {

	fileName := "Acroforms2.pdf"
	src := filepath.Join(inDir, fileName)
	src = "/Users/oneplus/Code/Work/gopdf/test/out/result1_by_parsed_ttf_font.pdf"
	dest := filepath.Join(outDir, fileName)

	lines := []api.Line{
		{
			X1: 0,
			Y1: 0,
			X2: 10,
			Y2: 10,
		},
	}

	err := api.DrawLineFile(src, dest, lines, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
