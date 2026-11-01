package analysis

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/klarerrs"
	"github.com/klarlanguage/klar/internal/module/imports"
	"github.com/klarlanguage/klar/internal/ranges"
	"github.com/klarlanguage/klar/internal/target"
)

// Importer resolves import paths to [Module]s.
type Importer interface {
	Import(imports.ImportPath, ImportContext) (*Module, error)
	// TODO: import wildcard that returns []*Module
}

type ImportContext struct {
	Target target.Target
	// The import path of the module that is importing.
	ImportPath imports.ImportPath
	// The folder path of the module that is importing.
	DirPath string
	// Whether the module that is importing is a single file.
	SingleFile bool
}

// performImports creates a new [Context] for each file and performs imports.
func (c *Checker) performImports(files []string, fileContexts map[string]*Context) {
	type imported struct {
		module *Module
		err    error
	}
	importCache := make(map[string]*imported)
	ictx := ImportContext{
		Target:     c.Module.Targets[0], // TODO
		ImportPath: c.Module.ImportPath,
		DirPath:    c.Module.Path,
		SingleFile: c.Module.Flags.Has(SingleFileModule),
	}
	for _, fileName := range files {
		fctx := fileContexts[fileName]
		var firstStmtI int
		// Perform imports for the file
		for i, stmt := range c.Programs[fileName].Body {
			imp, ok := stmt.(*ast.ImportStatement)
			if !ok {
				break
			}
			firstStmtI = i + 1

			impPathStr := imports.ImportPath.String(imp.Module)
			// Try to load from cache, or import it fresh and save it to cache
			res, ok := importCache[impPathStr]
			newImport := func() (res *imported) {
				res = &imported{}
				// If this module is a single file, report an error if the import isn't
				// from the stdlib. Don't call importModule because it would likely
				// report an error for the dependency not existing.
				if c.Module.Flags.Has(SingleFileModule) && !imports.ImportPath.IsStdlib(imp.Module) {
					res.err = klarerrs.ImportError(klarerrs.ErrSingleFileImport, imp.Module, nil)
					return
				}

				if res.module, res.err = c.importModule(imp.Module, ictx); res.err != nil {
					return
				}
				// Ensure the module is supported on the current targets
				if err := c.checkImportTargetSupport(res.module, imp); err != nil {
					// Setting res.err in the condition will set err to a nil *klarerrs.Error
					// rather than a nil `error`
					res.err = err
					return
				}
				// Ensure the module has any public exports, otherwise report an error
				if !res.module.HasExports() {
					res.err = klarerrs.ImportError(klarerrs.ErrNoPublicExports, imp.Module, nil)
				}
				return res
			}
			if !ok {
				res = newImport()
				importCache[impPathStr] = res
			}

			// Apply the import or report the error
			if res.err != nil {
				c.reportImportError(impPathStr, res.err, fctx.File, imp)
				c.declareErrorImport(imp, fctx)
				continue
			}
			c.applyImportedModule(res.module, imp, fctx)
		}
		fctx.setAttribute(firstStmtIndex, firstStmtI)
	}

	// Populate [Module.Imports]
	// TODO: Should modules with errors be included?
	c.Module.Imports = make([]*Module, 0, len(importCache))
	for _, importPath := range slices.Sorted(maps.Keys(importCache)) {
		mod := importCache[importPath].module
		if mod != nil {
			c.Module.Imports = append(c.Module.Imports, mod)
		}
	}
}

func (c *Checker) importModule(p imports.ImportPath, ictx ImportContext) (*Module, error) {
	if c.Options.Importer == nil {
		// Importer not set up, meaning imports aren't allowed
		return nil, &klarerrs.Error{Code: klarerrs.ErrImporterNotFound}
	}
	return c.Options.Importer.Import(p, ictx)
}

func (c *Checker) applyImportedModule(mod *Module, stmt *ast.ImportStatement, fctx *Context) {
	if stmt.Alias != nil && stmt.Alias.IsDiscard() {
		return // Don't do anything
	}
	if stmt.Wildcard {
		// We have to import the submodules
	}
	ns := mod.ImportPath.Namespace()
	if stmt.Alias != nil {
		ns = stmt.Alias.Name
	}

	// Declare the namespace
	nsType := &Namespace{ImportPath: mod.ImportPathString(), Context: mod.Context}
	nsObj := NewObject(ns, fctx.File, stmt.Range, c.Module, nsType)
	// The namespace doesn't have to be used when there are unqualified imports
	if len(stmt.UnqualifiedImports) > 0 {
		nsObj.Flags |= UsageOptional
	}
	c.declare(fctx, nsObj)

	// Declare unqualified imports
	for _, name := range stmt.UnqualifiedImports {
		target := name.Name.Name
		obj, err := nsType.lookupExport(target)
		if err != nil {
			err.Range = name.Name.Range()
			c.fileError(err, fctx.File)
			continue
		}
		// Clone the object so
		// - We can modify its declared range, so when an error is reported for
		// 	its name being redeclared, it shows the range of the import rather
		// 	than the definition outside the current module.
		// - We can use the user-declared unqualified import alias
		obj = obj.Clone(c.Module, fctx.File, name.Name.Range())
		// TODO: Should we change obj's order, and context?
		if !name.Label.IsZero() {
			obj.Name = name.Label.Name
		}
		obj.Public = false // To ensure warnings are reported if unused
		// Declare the unqualified import
		// TODO: Use a custom redeclared error to show "X was already imported"
		c.declare(fctx, obj)
	}
}

func (c *Checker) declareErrorImport(stmt *ast.ImportStatement, fctx *Context) {
	if stmt.Alias != nil && stmt.Alias.IsDiscard() {
		return // Don't do anything
	}
	ns := imports.ImportPath.Namespace(stmt.Module)
	if stmt.Alias != nil {
		ns = stmt.Alias.Name
	}
	// Declare the namespace
	nsObj := NewObject(ns, fctx.File, stmt.Range, c.Module, &InvalidObject{})
	c.declare(fctx, nsObj)
}

func (c *Checker) reportImportError(importPath string, err error,
	fid FileID, stmt *ast.ImportStatement,
) {
	kerr, ok := err.(*klarerrs.Error)
	if ok {
		kerr = new(*kerr) // Copy the importer's error so we can add location info
	} else {
		kerr = &klarerrs.Error{
			Code: klarerrs.ErrImporterError,
			Info: klarerrs.ModuleErrorInfo{ImportPath: importPath, Err: err},
		}
	}
	kerr.Node = stmt
	kerr.Range = stmt.Range

	// Helpful error label
	importPathBase, _, _ := strings.Cut(importPath, ".")
	quotedImportPath := klarerrs.Quote(importPath)
	switch {
	case kerr.Code == klarerrs.ErrNoPublicExports:
		kerr.Label = quotedImportPath + " doesn't provide anything to import"
	case kerr.Code != klarerrs.ErrModuleNotFound:
		// Keep the provided label or use no label
	case slices.Contains(imports.StdlibImports, importPathBase):
		kerr.Label = quotedImportPath + " isn't in the standard library"
	case importPathBase == c.Module.ImportPath[0]:
		kerr.Label = quotedImportPath + " isn't in the current package"
	default:
		kerr.Label = "Can't find " + quotedImportPath
	}

	c.fileError(kerr, fid)
}

func (c *Checker) checkImportTargetSupport(imported *Module, stmt *ast.ImportStatement) *klarerrs.Error {
	for _, currTarget := range c.Module.Targets {
		if target.Supports(imported.Targets, currTarget) {
			continue
		}
		err := klarerrs.Node(klarerrs.ErrUnsupportedImportTarget, stmt).
			SetParam("supported", imported.Targets)
		err.Name = currTarget.Name()
		err.Info = klarerrs.ModuleErrorInfo{ImportPath: imported.ImportPathString()}

		var b strings.Builder
		b.WriteString("The module supports these targets:\n  ")
		for i, t := range imported.Targets {
			if t == target.Unknown {
				continue
			}
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t.Name())
		}
		err.Desc = b.String()
		return err
	}
	return nil
}

type Namespace struct {
	ImportPath string
	Context    *Context
}

func (*Namespace) Kind() Kind          { return KindNamespace }
func (ns *Namespace) Underlying() Type { return ns }
func (*Namespace) objKind()            {}
func (ns *Namespace) String() string   { return "module " + ns.ImportPath }

func (ns *Namespace) lookupExport(target string) (*Object, *klarerrs.Error) {
	obj := ns.Context.Lookup(target)
	if obj == nil || !obj.Public {
		// Not found in the module or private
		err := klarerrs.ReferenceError(
			klarerrs.ErrExportUndefined, ranges.Range{}, target,
		).SetParam("module", ns.ImportPath)
		err.Label = fmt.Sprintf(
			"%s doesn't export %s",
			klarerrs.Quote(ns.ImportPath), klarerrs.Quote(target),
		)
		if obj != nil {
			err.Code = klarerrs.ErrNotExported
			err.AddDetail(
				klarerrs.Quote(target)+" was declared here, and it isn't public",
				obj.FilePath(), obj.Range,
			)
		}
		return nil, err
	}
	return obj, nil
}

func (ns *Namespace) Index(field string, t *Expr) (err *klarerrs.Error) {
	t.Type, err = ns.lookupExport(field)
	return err
}
