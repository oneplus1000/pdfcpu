package pdfcpu

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/types"
	"github.com/pkg/errors"
)

func AnalyzeContext(ctx *Context) (*types.AnalyzeResult, error) {
	var result types.AnalyzeResult
	err := analyze(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return nil, nil
}

func analyze(ctx *Context, result *types.AnalyzeResult) error {

	// Get a reference to the PDF indirect reference of the page tree root dict.
	indRefPages, err := ctx.Pages()
	if err != nil {
		return fmt.Errorf("ctx.Pages() fail %w", err)
	}

	// Dereference and get a reference to the page tree root dict.
	pageTreeRootDict, err := ctx.XRefTable.DereferenceDict(*indRefPages)
	if err != nil {
		return fmt.Errorf("ctx.XRefTable.DereferenceDict(*indRefPages) fail %w", err)
	}

	// Detect the number of pages of this PDF file.
	pageCount := pageTreeRootDict.IntEntry("Count")
	if pageCount == nil {
		return errors.New("pdfcpu: optimizeFontAndImagess: missing \"Count\" in page root dict")
	}

	_, err = analyzePagesDict(ctx, pageTreeRootDict, 0)
	if err != nil {
		return fmt.Errorf("analyzePagesDict(ctx, pageTreeRootDict, 0) fail %w", err)
	}

	return nil
}

func analyzePagesDict(ctx *Context, pagesDict Dict, pageNumber int) (int, error) {

	// Iterate over page tree.
	kids := pagesDict.ArrayEntry("Kids")
	for _, v := range kids {

		// Dereference next page node dict.
		ir, _ := v.(IndirectRef)
		o, err := ctx.Dereference(ir)
		if err != nil {
			return 0, fmt.Errorf("parsePagesDict: can't locate Pagedict or Pagesdict %w", err)
		}

		pageNodeDict := o.(Dict)
		dictType := pageNodeDict.Type()
		if dictType == nil {
			return 0, errors.New("pdfcpu: parsePagesDict: Missing dict type")
		}

		if *dictType == "Pages" {

			// Recurse over pagetree and optimize resources.
			pageNumber, err = analyzePagesDict(ctx, pageNodeDict, pageNumber)
			if err != nil {
				return 0, err
			}

			continue
		}

		if *dictType != "Page" {
			return 0, errors.Errorf("pdfcpu: parsePagesDict: Unexpected dict type: %s\n", *dictType)
		}

		// Parse and optimize resource dict for one page.
		if err = analyzeResourcesDict(ctx, pageNodeDict, pageNumber, int(ir.ObjectNumber)); err != nil {
			return 0, err
		}

		//fmt.Printf("pageNumber=%d\n", pageNumber)
		pageNumber++
	}

	return pageNumber, nil
}

func analyzeResourcesDict(ctx *Context, pageDict Dict, pageNumber, pageObjNumber int) error {

	// Get resources dict for this page.
	d, err := resourcesDictForPageDict(ctx.XRefTable, pageDict, pageObjNumber)
	if err != nil {
		return err
	}

	if d != nil {
		err := analyzeResources(ctx, d, pageNumber, pageObjNumber)
		if err != nil {
			return fmt.Errorf("analyzeResources(...) fail %w", err)
		}
	}

	return nil
}

func analyzeResources(ctx *Context, resourcesDict Dict, pageNumber, pageObjNumber int) error {
	o, found := resourcesDict.Find("XObject")
	if found {
		d, err := ctx.DereferenceDict(o)
		if err != nil {
			return fmt.Errorf("ctx.DereferenceDict(o) fail %w", err)
		}
		if d == nil {
			return errors.Errorf("pdfcpu: optimizeResources: xobject resource dict is null for page %d pageObj %d\n", pageNumber, pageObjNumber)
		}

		err = analyzeXObjectResourcesDict(ctx, d, pageNumber, pageObjNumber)
		if err != nil {
			return fmt.Errorf("analyzeXObjectResourcesDict(...) fail %w", err)
		}
	}
	return nil
}

func analyzeXObjectResourcesDict(ctx *Context, rDict Dict, pageNumber, pageObjNumber int) error {

	//pageImages := pageImages(ctx, pageNumber)
	for rName, v := range rDict {
		indRef, ok := v.(IndirectRef)
		if !ok {
			return errors.Errorf("pdfcpu: optimizeXObjectResourcesDict: missing indirect object ref for resourceId: %s", rName)
		}
		objNr := int(indRef.ObjectNumber)
		entry, found := ctx.FindTableEntry(indRef.ObjectNumber.Value(), indRef.GenerationNumber.Value())
		if !found {
			return nil //do noting
		}
		entry.Valid = true
		osd, ok := entry.Object.(StreamDict)
		if !ok {
			return errors.Errorf("pdfcpu: dereferenceStreamDict: wrong type <%v> %T", o, entry.Object)
		}

		if osd.Dict.Subtype() == nil {
			return errors.Errorf("pdfcpu: optimizeXObjectResourcesDict: missing stream dict Subtype %s\n", v)
		} else if *osd.Dict.Subtype() == "Image" {
			//fmt.Printf("img %d pageNumber = %d\n", objNr, pageNumber)
			return analyzeImage(ctx, osd, pageNumber, objNr)
		}
	}
	return nil
}

func analyzeImage(ctx *Context, stmDict StreamDict, pageNumber, pageObjNumber int) error {
	//fmt.Printf("\tstmDict.StreamLength %+v\n", (stmDict.Dict))
	_, err := WriteImage(ctx.XRefTable, fmt.Sprintf("/Users/oneplus/Desktop/m/xxx_%d_%d", pageNumber, pageObjNumber), &stmDict, pageObjNumber)
	if err != nil {
		fmt.Printf("err:%+v\n", err)
	}
	return nil
}
