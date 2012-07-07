package cxxgo

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sbinet/go-cxxdict/pkg/cxxtypes"
	"github.com/sbinet/go-cxxdict/pkg/wrapper"
)

type bufmap_t map[string]*bytes.Buffer

type plugin struct {
	gen *wrapper.Generator // the generator which is invoking us
	sel []string           // the identifiers to select
	ids []string           // the selected identifiers
}

func (p *plugin) Name() string {
	return "cxxgo.plugin"
}

func (p *plugin) Init(g *wrapper.Generator) error {
	if g == nil {
		return fmt.Errorf("cxxgo: nil pointer to wrapper.Generator")
	}
	p.gen = g
	p.sel = []string{
		"T*",
		"Math*",
		"NS::*tmpl*",
		"Foo",
		"IFoo",
		"NS*",
		"TT*",
		"Base*",
		"Alg*",
		"With*Base*",

		"*Enum*",
		"cblas*",
		"*CBLAS_*",
	}
	p.ids = []string{}

	fmt.Printf("cxxgo.Init: args=%v\n", g.Args)
	return nil
}

func (p *plugin) Generate(fd *wrapper.FileDescriptor) error {
	fmt.Printf("cxxgo.Generate...\n")

	// loop over identifiers and filter them out
	for _, n := range cxxtypes.IdNames() {
		selected := false
		for _, sel := range p.sel {
			matched, err := path.Match(sel, n)
			if err != path.ErrBadPattern && matched {
				selected = true
				break
			}
		}
		if selected && is_anon(n) {
			fmt.Printf(":: discarding [%s] (anonymous identifier)\n", n)
			selected = false
		}
		if selected {
			p.ids = append(p.ids, n)
			nn := gen_go_name_from_id(cxxtypes.IdByName(n))
			_cxx2go_typemap[n] = nn
		}
	}
	{
		// make sure we don't wrap a member twice: remove duplicates
		nremoved := 0
		for {
			nremoved = 0
			//fmt.Printf("---\n")
			sel_ids := make([]string, 0, len(p.ids))
			for _, n := range p.ids {
				id := cxxtypes.IdByName(n)
				if mbr, ok := id.(*cxxtypes.Member); ok {
					pid := cxxtypes.IdByName(mbr.Scope)
					if pid != nil &&
						str_is_in_slice(pid.IdScopedName(), p.ids) {
						// parent is already selected... discard member
						nremoved += 1
						//fmt.Printf("** discard [%s]\n", n)
					} else {
						// keep it
						sel_ids = append(sel_ids, n)
					}
				} else {
					sel_ids = append(sel_ids, n)
				}
			}
			p.ids = sel_ids
			if nremoved == 0 {
				break
			}
		}
	}
	fmt.Printf("selected ids: %v\n", p.ids)
	if len(p.ids) <= 0 {
		fmt.Printf("nothing to wrap\n")
		return nil
	}

	fd_cxx, err := os.Create(fd.Name + "_" + p.Name() + ".cxx")
	if err != nil {
		return err
	}
	defer fd_cxx.Close()

	fd_hdr, err := os.Create(fd.Name + "_" + p.Name() + ".h")
	if err != nil {
		return err
	}
	defer fd_hdr.Close()

	fd_go, err := os.Create(fd.Name + "_" + p.Name() + ".go")
	if err != nil {
		return err
	}
	defer fd_go.Close()

	fd.Files["cxx"] = fd_cxx
	fd.Files["hdr"] = fd_hdr
	fd.Files["go"] = fd_go

	_, err = fd_go.WriteString(fmt.Sprintf(
		_go_hdr,
		fd.Package,
		fd_hdr.Name(),
		fd.Name,
	))
	if err != nil {
		return err
	}

	_, err = fd_cxx.WriteString(fmt.Sprintf(
		_cxx_hdr,
		fd_hdr.Name(),
		fd.Header,
	))
	if err != nil {
		return err
	}

	_, err = fd_hdr.WriteString(fmt.Sprintf(
		_hdr_hdr,
		fd.Package,
		fd.Package,
	))
	if err != nil {
		return err
	}

	for _, n := range p.ids {
		id := cxxtypes.IdByName(n)
		switch id := id.(type) {
		case *cxxtypes.ClassType:
			err := p.wrapClass(id)
			if err != nil {
				return err
			}

		case *cxxtypes.StructType:
			err := p.wrapStruct(id)
			if err != nil {
				return err
			}

		case *cxxtypes.OverloadFunctionSet:
			err := p.wrapFunction(id)
			if err != nil {
				return err
			}

		case *cxxtypes.Namespace:
			err := p.wrapNamespace(id)
			if err != nil {
				return err
			}

		case *cxxtypes.Member:
			// will be done by a scope-level thingy...

			// err := p.wrapMember(id)
			// if err != nil {
			// 	return err
			// }

		case *cxxtypes.TypedefType:
			err := p.wrapTypedef(id)
			if err != nil {
				return err
			}

		case *cxxtypes.RefType:
			err := p.wrapRefType(id)
			if err != nil {
				return err
			}

		case *cxxtypes.PtrType:
			err := p.wrapPtrType(id)
			if err != nil {
				return err
			}

		case *cxxtypes.CvrQualType:
			err := p.wrapCvrQualType(id)
			if err != nil {
				return err
			}

		case *cxxtypes.EnumType:
			err := p.wrapEnum(id)
			if err != nil {
				return err
			}

		default:
			panic(fmt.Errorf("type [%T] unhandled (%s)!", id, id.IdScopedName()))
		}
	}

	_, err = fd_go.WriteString(fmt.Sprintf(
		_go_footer,
		fd.Package,
	))
	if err != nil {
		return err
	}

	_, err = fd_cxx.WriteString(_cxx_footer)
	if err != nil {
		return err
	}

	_, err = fd_hdr.WriteString(fmt.Sprintf(
		_hdr_footer,
		fd.Package,
	))
	if err != nil {
		return err
	}

	fmt.Printf("cxxgo.Generate... [done]\n")
	return nil
}

func (p *plugin) wrapClass(id *cxxtypes.ClassType) (err error) {
	fmt.Printf(":: wrapping class [%s]...\n", id.IdScopedName())
	err = nil

	bufs := new_bufmap("cxx_head",
		"cxx_body",
		"go_impl",
		"go_iface",
	)

	clf := "::" + id.IdScopedName()
	clt := g_strtrans.Replace(id.IdScopedName())
	go_cls_iface_name := gen_go_name_from_id(id)
	go_cls_impl_name := "Gocxxcptr" + go_cls_iface_name

	fmt.Printf("==> %s => %s\n", id.IdScopedName(), clt)

	fmter(bufs["go_iface"],
		`
// %s wraps the C++ class %s
type %s interface {
	Gocxxcptr() uintptr
	GocxxIs%s()
`,
		go_cls_iface_name,
		clf,
		go_cls_iface_name,
		go_cls_iface_name,
	)

	fmter(bufs["go_impl"], "type %s uintptr\n", go_cls_impl_name)

	// bases...
	bufs_bases := make([]bufmap_t, 0, len(id.Bases))
	for _, base := range id.Bases {
		bufs_base := bufmap_t{
			"go_impl":  bytes.NewBufferString(""),
			"go_iface": bytes.NewBufferString(""),
			"cxx":      bytes.NewBufferString(""),
		}
		bufs_bases = append(bufs_bases, bufs_base)
		err := p.wrapBaseClass(&base, bufs)
		if err != nil {
			return err
		}
		if base.IsPublic() {
			bid := base.Type().(cxxtypes.Id)
			go_base_cls_iface_name := gen_go_name_from_id(bid)
			fmter(bufs["go_iface"],
				"\tGet%s() %s\n",
				go_base_cls_iface_name,
				go_base_cls_iface_name)
		}
	}

	fct_mbr_indices := make([]int, 0, len(id.Members))
	fct_mbr_names := make([]string, 0, len(id.Members))
	// data-members
	for i, mbr := range id.Members {
		if !p.mbr_filter(&mbr) {
			fmt.Printf(":: discarding [%s]...\n", mbr.Name)
			continue
		}
		if mbr.IsFunctionMember() {
			//FIXME: O(n^2)
			if !str_is_in_slice(mbr.Name, fct_mbr_names) {
				fct_mbr_names = append(fct_mbr_names, mbr.Name)
				fct_mbr_indices = append(fct_mbr_indices, i)
			}
			continue
		}
		mid := cxxtypes.IdByName(mbr.Name)
		if mid == nil {
			fmt.Printf("==[%s]==(idx=%d)\n", mbr.Name, i)
			fmt.Printf("==dmbr: %v\n", mbr.IsDataMember())
			fmt.Printf("==fmbr: %v\n", mbr.IsFunctionMember())
			fmt.Printf("==embr: %v\n", mbr.IsEnumMember())
			fmt.Printf("==mkind: %v\n", mbr.Kind)
			fmt.Printf("==mdind: %v\n", mbr.IdKind())
			return fmt.Errorf("cxxgo: could not retrieve identifier [%s]\n%s", mbr.Name, mbr)
		}
		fmt.Printf("--> (%s)[%s]...\n", mbr.IdScopedName(), mbr)
		err := p.wrapMember(&mbr, bufs)
		if err != nil {
			return err
		}
	}

	// fct-members: handle overloads...
	for _, mbr_idx := range fct_mbr_indices {
		mbr := &id.Members[mbr_idx]
		err := p.wrapFctMember(mbr, bufs)
		if err != nil {
			return err
		}
	}

	fmter(bufs["go_iface"], "}\n\n")

	// commit buffers
	_, err = bufs["go_iface"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	_, err = bufs["go_impl"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	fmt.Printf(":: wrapping class [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapStruct(id *cxxtypes.StructType) error {
	fmt.Printf(":: wrapping struct [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping struct [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapFunction(id *cxxtypes.OverloadFunctionSet) error {
	fmt.Printf(":: wrapping fct [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping fct [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapNamespace(id *cxxtypes.Namespace) error {
	fmt.Printf(":: wrapping namespace [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping namespace [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapEnum(id *cxxtypes.EnumType) error {
	var err error = nil
	fmt.Printf(":: wrapping enum [%s]...\n", id.IdScopedName())

	n := "::" + id.IdScopedName()
	//tn := g_strtrans.Replace(id.IdScopedName())
	go_enum_iface_name := gen_go_name_from_id(id)

	bufs := new_bufmap("cxx_head",
		"go_iface",
	)

	fmter(bufs["go_iface"],
		"\n// %s wraps the enum %s\n", go_enum_iface_name, n)
	fmter(bufs["go_iface"], "type %s int\n", go_enum_iface_name)

	// commit buffers
	_, err = bufs["go_iface"].WriteTo(p.gen.Fd.Files["go"])
	if err != nil {
		return err
	}

	fmt.Printf(":: wrapping enum [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapMember(id *cxxtypes.Member, bufs bufmap_t) error {
	fmt.Printf(":: wrapping member [%s]...\n", id.IdScopedName())
	if id.IsDataMember() {
		err := p.wrapDataMember(id, bufs)
		if err != nil {
			return err
		}
	} else if id.IsFunctionMember() {
		err := p.wrapFctMember(id, bufs)
		if err != nil {
			return err
		}
	}
	fmt.Printf(":: wrapping member [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapTypedef(id *cxxtypes.TypedefType) error {
	fmt.Printf(":: wrapping typedef [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping typedef [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapRefType(id *cxxtypes.RefType) error {
	fmt.Printf(":: wrapping ref-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping ref-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapPtrType(id *cxxtypes.PtrType) error {
	fmt.Printf(":: wrapping ptr-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping ptr-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapCvrQualType(id *cxxtypes.CvrQualType) error {
	fmt.Printf(":: wrapping cvr-type [%s]...\n", id.IdScopedName())

	fmt.Printf(":: wrapping cvr-type [%s]...[ok]\n", id.IdScopedName())
	return nil
}

func (p *plugin) wrapBaseClass(id *cxxtypes.Base, bufs bufmap_t) error {
	return nil
}

func (p *plugin) wrapDataMember(id *cxxtypes.Member, bufs bufmap_t) (err error) {
	fmt.Printf(":: wrapping data-member [%s]...\n", id.IdScopedName())

	dm_name := id.IdName()
	dm_typename := id.Type

	// declare getters and setters in the go-interface
	fmter(bufs["go_iface"], "\tGet%s() %s\n\tSet%s(%s %s)\n",
		strings.Title(dm_name),
		cxx2go_typename(dm_typename),
		strings.Title(dm_name),
		dm_name,
		cxx2go_typename(dm_typename),
	)

	clsid := cxxtypes.IdByName(id.Scope)
	if clsid == nil {
		return fmt.Errorf("could not find parent-scope [%s] for member [%s]",
			id.Scope, id.IdScopedName())
	}
	//clf := "::" + clsid.IdScopedName()
	//clt := g_strtrans.Replace(clsid.IdScopedName())
	go_cls_iface_name := gen_go_name_from_id(clsid)
	go_cls_impl_name := "Gocxxcptr" + go_cls_iface_name

	// corresponding implementation...
	fmter(bufs["go_impl"],
		`
func (p %s) Get%s() %s {
 var dummy %s
 return dummy
}
`,
		go_cls_impl_name,
		strings.Title(dm_name),
		cxx2go_typename(dm_typename),
		cxx2go_typename(dm_typename),
	)

	fmter(bufs["go_impl"],
		`
func (p %s) Set%s(arg %s) {
}
`,
		go_cls_impl_name,
		strings.Title(dm_name),
		cxx2go_typename(dm_typename),
	)
	fmt.Printf(":: wrapping data-member [%s]...[ok]\n", id.IdScopedName())
	return err
}

func (p *plugin) wrapFctMember(id *cxxtypes.Member, bufs bufmap_t) error {
	fmt.Printf(":: wrapping fct-member [%s]...\n", id.IdScopedName())
	ovfct := cxxtypes.IdByName(id.Name).(*cxxtypes.OverloadFunctionSet)
	ndargs := 0
	dbg := false
	if id.IdScopedName() == "Foo::setFoo" {
		dbg = true
	}
	for i, _ := range ovfct.Fcts {
		if ovfct.Function(i).NumDefaultParam() > 0 {
			ndargs += 1
		}
		if dbg {
			fmt.Printf("== ovfct[%d]...\n", i)
			for ii, pp := range ovfct.Fcts[i].Params {
				fmt.Printf("arg[%d]{%s, %s %v}\n", ii, pp.Name, pp.Type, pp.DefVal)
			}
		}
		fmt.Printf(" --> [%s]\n", get_prototype(ovfct.Fcts[i]))
	}
	fmt.Printf("   #overloads: %d\n", ovfct.NumFunction())
	fmt.Printf("   #dflt-args: %d\n", ndargs)
	fmt.Printf(":: wrapping fct-member [%s]...[ok]\n", id.IdScopedName())
	return nil
}

// utils ------------------------

func fmter(buf *bytes.Buffer, format string, args ...interface{}) (int, error) {
	o := fmt.Sprintf(format, args...)
	return buf.WriteString(o)
}

func new_bufmap(keys ...string) bufmap_t {
	bufs := make(bufmap_t, len(keys))
	for _, k := range keys {
		bufs[k] = bytes.NewBufferString("")
	}
	return bufs
}

// cxx2go_typename converts a C++ type string into its go equivalent
func cxx2go_typename(t string) string {
	if o, ok := _cxx2go_typemap[t]; ok {
		return o
	}
	o := fmt.Sprintf("_go_unknown_%s", t)
	_cxx2go_typemap[t] = o
	return o
}

// cxx2cgo_typemap converts a C++ type string into its cgo equivalent
func cxx2cgo_typename(t string) string {
	return cxx2go_typename(t)
}

func gen_go_name_from_id(id cxxtypes.Id) string {
	n := id.IdScopedName()

	// special cases
	if _, ok := _cxx2go_typemap[n]; ok {
		return cxx2go_typename(n)
	}

	switch id := id.(type) {

	case *cxxtypes.Function, *cxxtypes.OverloadFunctionSet,
		*cxxtypes.ClassType:
		n = strings.Title(n)

	case *cxxtypes.RefType:
		return gen_go_name_from_id(id.UnderlyingType().(cxxtypes.Id))

	case *cxxtypes.CvrQualType:
		return gen_go_name_from_id(cxxtypes.IdByName(id.Type))
	}

	// sanitize
	o := g_cxxgo_trans.Replace(n)

	if _, ok := _cxx2go_typemap[o]; ok {
		return cxx2go_typename(o)
	}
	return o
}

func gen_go_name(cxxname string) string {
	o := g_cxxgo_trans.Replace(cxxname)
	if _, ok := _cxx2go_typemap[o]; ok {
		return cxx2go_typename(o)
	}
	return o
}

func (p *plugin) mbr_filter(mbr *cxxtypes.Member) bool {
	if mbr == nil {
		return false
	}

	// filter out any non public member
	if mbr.IsPrivate() || mbr.IsProtected() {
		return false
	}

	// filter out any anonymous member
	if n := mbr.Name; is_anon(n) {
		return false
	}

	// filter any copy constructor with a private copy constructor in
	// any base class
	// TODO

	// filter any constructor for pure abstract classes
	// TODO

	// filter methods taking non-public args
	// TODO

	// filter using the exclusion list in the selection file
	// TODO

	return true
}

// get_container_id returns the container-type and stl-class of id
// (if id is actually a container.)
func get_container_id(id cxxtypes.Id) (string, string) {
	n := id.IdScopedName()
	if strings.HasSuffix(n, "iterator") {
		return "NOCONTAINER", ""
	} else if strings.HasSuffix(n, "iterator<true>") {
		return "NOCONTAINER", ""
	} else if strings.HasSuffix(n, "iterator<false>") {
		return "NOCONTAINER", ""
	} else if strings.HasPrefix(n, "std::deque") {
		return "DEQUE", "list"
	} else if strings.HasPrefix(n, "std::list") {
		return "LIST", "list"
	} else if strings.HasPrefix(n, "std::map") {
		return "MAP", "map"
	} else if strings.HasPrefix(n, "std::multimap") {
		return "MULTIMAP", "map"
	} else if strings.HasPrefix(n, "std::unordered_map") {
		return "HASHMAP", "map"
	} else if strings.HasPrefix(n, "std::unordered_multimap") {
		return "HASHMULTIMAP", "map"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_map") {
		return "HASHMAP", "map"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_multimap") {
		return "HASHMULTIMAP", "map"
	} else if strings.HasPrefix(n, "stdext::hash_map") {
		return "HASHMAP", "map"
	} else if strings.HasPrefix(n, "stdext::hash_multimap") {
		return "HASHMULTIMAP", "map"
	} else if strings.HasPrefix(n, "std::queue") {
		return "QUEUE", "queue"
	} else if strings.HasPrefix(n, "std::set") {
		return "SET", "set"
	} else if strings.HasPrefix(n, "std::multiset") {
		return "MULTISET", "set"
	} else if strings.HasPrefix(n, "std::unordered_set") {
		return "HASHSET", "set"
	} else if strings.HasPrefix(n, "std::unordered_multiset") {
		return "HASHMULTISET", "set"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_set") {
		return "HASHSET", "set"
	} else if strings.HasPrefix(n, "__gnu_cxx::hash_multiset") {
		return "HASHMULTISET", "set"
	} else if strings.HasPrefix(n, "stdext::hash_set") {
		return "HASHSET", "set"
	} else if strings.HasPrefix(n, "stdext::hash_multiset") {
		return "HASHMULTISET", "set"
	} else if strings.HasPrefix(n, "std::stack") {
		return "STACK", "stack"
	} else if strings.HasPrefix(n, "std::vector") {
		return "VECTOR", "vector"
	} else if strings.HasPrefix(n, "std::bitset") {
		return "BITSET", "bitset"
	} else {
		return "NOCONTAINER", ""
	}
	panic("unreachable")
}

// is_anon returns true if the given typename looks like an anonymous one
func is_anon(n string) bool {
	// ellipsis would screw us up...
	nn := strings.Replace(n, "...", "", -1)
	if strings.IndexAny(nn, ".$") != -1 {
		return true
	}
	return false
}

func str_is_in_slice(s string, slice []string) bool {
	for _, ss := range slice {
		if s == ss {
			return true
		}
	}
	return false
}

func get_prototype(fct *cxxtypes.Function) string {
	s := []string{}
	if fct.IsInline() {
		s = append(s, "inline ")
	}
	if fct.IsStatic() {
		s = append(s, "static ")
	}
	//fixme: add fct qualifiers: const|static|inline
	if !fct.IsConstructor() && !fct.IsDestructor() {
		s = append(s, fct.ReturnType().TypeName(), " ")
	}
	s = append(s, fct.IdScopedName(), "(")
	if len(fct.Params) > 0 {
		for i, _ := range fct.Params {
			s = append(s,
				strings.TrimSpace(fct.Param(i).Type),
				" ",
				strings.TrimSpace(fct.Param(i).Name))
			if i < len(fct.Params)-1 {
				s = append(s, ", ")
			}
		}
	} else {
		// fixme: we should rather test if C XOR C++...
		if fct.IsMethod() {
			//nothing
		} else {
			s = append(s, "void")
		}
	}
	if fct.IsVariadic() {
		s = append(s, "...")
	}
	s = append(s, ") ")
	if fct.IsConst() {
		s = append(s, "const ")
	}
	return strings.TrimSpace(strings.Join(s, ""))
}

// globals ----------------------

var g_strtrans *strings.Replacer = strings.NewReplacer(
	"<", "_",
	">", "_",
	"&", "r",
	"*", "p",
	",", "_",
	":", "_",
	" ", "s",
	"(", "_",
	")", "_",
	".", "_",
	"$", "d",
	"-", "m",
	"[", "_",
	"]", "_",
)

var g_cxxgo_trans *strings.Replacer = strings.NewReplacer(
	"<", "_Sl_",
	">", "_Sg_",
	",", "_Sc_",
	" ", "_",
	"-", "m",
	"::", "_",
)

var _go_hdr string = `
package %s

// #include <stdlib.h>
// #include <string.h>
// #include "%s"
// #cgo LDFLAGS: -l%s
import "C"
import "unsafe"

// dummy function which uses unsafe
func _gocxx_free_ptr(ptr unsafe.Pointer) {
 C.free(ptr)
}
`

var _cxx_hdr string = `
// C includes
#include <stdlib.h>
#include <string.h>

// C++ includes
#include <string>
#include <vector>

#include "%s"

#include "%s"

#ifdef __cplusplus
extern "C" {
#endif

// helpers for CGo runtime

typedef struct { char *p; int n; } _gostring_;
typedef struct { void* array; unsigned int len; unsigned int cap; } _goslice_;


extern void crosscall2(void (*fn)(void *, int), void *, int);
extern void _cgo_allocate(void *, int);
extern void _cgo_panic(void *, int);

static void *_gocxx_goallocate(size_t len) {
  struct {
    size_t len;
    void *ret;
  } a;
  a.len = len;
  crosscall2(_cgo_allocate, &a, (int) sizeof a);
  return a.ret;
}

static void _gocxx_gopanic(const char *p) {
  struct {
    const char *p;
  } a;
  a.p = p;
  crosscall2(_cgo_panic, &a, (int) sizeof a);
}

static _gostring_ _gocxx_makegostring(const char *p, size_t l) {
  _gostring_ ret;
  ret.p = (char*)_gocxx_goallocate(l + 1);
  memcpy(ret.p, p, l);
  ret.n = l;
  return ret;
}

#define GOCXX_contract_assert(expr, msg) \
  if (!(expr)) { _gocxx_gopanic(msg); } else

#define GOCXX_exception(code, msg) _gocxx_gopanic(msg)
`

var _hdr_hdr string = `
#ifndef _GOCXXDICT_%s_H
#define _GOCXXDICT_%s_H 1

#ifdef __cplusplus
extern "C" {
#endif
`

var _go_footer string = `
// EOF %s
`

var _cxx_footer string = `
#ifdef __cpluspluc
} /* extern "C" */
#endif
`

var _hdr_footer string = `
#ifdef __cplusplus
} /* extern "C" */
#endif

#endif /* ! %s_H */
`

var _cxx2go_typemap = map[string]string{
	"void":     "",
	"uint64_t": "uint64",
	"uint32_t": "uint32",
	"uint16_t": "uint16",
	"uint8_t":  "uint8",
	"uint_t":   "uint",
	"int64_t":  "int64",
	"int32_t":  "int32",
	"int16_t":  "int16",
	"int8_t":   "int8",

	"bool":           "bool",
	"char":           "byte",
	"signed char":    "int8",
	"unsigned char":  "byte",
	"short":          "int16",
	"unsigned short": "uint16",
	"int":            "int",
	"unsigned int":   "uint",

	// FIXME: 32/64 platforms... (and cross-compilation)
	//"long":           "int32",
	//"unsigned long":  "uint32",
	"long":          "int64",
	"unsigned long": "uint64",

	"long long":          "int64",
	"unsigned long long": "uint64",

	"float":  "float32",
	"double": "float64",

	"float complex":  "complex64",
	"double complex": "complex128",

	// FIXME: 32/64 platforms
	//"size_t": "int",
	"size_t": "int64",

	// stl
	"std::string": "string",

	// ROOT types
	"Char_t":   "byte",
	"UChar_t":  "byte",
	"Short_t":  "int16",
	"UShort_t": "uint16",
	"Int_t":    "int",
	"UInt_t":   "uint",

	"Seek_t":     "int",
	"Long_t":     "int64",
	"ULong_t":    "uint64",
	"Float_t":    "float32",
	"Float16_t":  "float32", //FIXME
	"Double_t":   "float64",
	"Double32_t": "float64",

	"Bool_t":    "bool",
	"Text_t":    "byte",
	"Byte_t":    "byte",
	"Version_t": "int16",
	"Option_t":  "byte",
	"Ssiz_t":    "int",
	"Real_t":    "float32",
	"Long64_t":  "int64",
	"ULong64_t": "uint64",
	"Axis_t":    "float64",
	"Stat_t":    "float64",
	"Font_t":    "int16",
	"Style_t":   "int16",
	"Marker_t":  "int16",
	"Width_t":   "int16",
	"Color_t":   "int16",
	"SCoord_t":  "int16",
	"Coord_t":   "float64",
	"Angle_t":   "float32",
	"Size_t":    "float32",
}

func init() {
	wrapper.RegisterPlugin(&plugin{})
}

// test interfaces...

var _ wrapper.Plugin = (*plugin)(nil)

// EOF