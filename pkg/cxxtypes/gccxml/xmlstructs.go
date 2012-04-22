package gccxml

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// helper function to "convert" a 0|1 string into a boolean
func to_bool(v string) bool {
	if v == "" || v == "0" {
		return false
	}
	return true
}

type xmlTree struct {
	XMLName xml.Name `xml:"GCC_XML"`

	Arrays            []*xmlArray            `xml:"ArrayType"`
	Classes           []*xmlClass            `xml:"Class"`
	Constructors      []*xmlConstructor      `xml:"Constructor"`
	Converters        []*xmlConverter        `xml:"Converter"`
	CvQualifiedTypes  []*xmlCvQualifiedType  `xml:"CvQualifiedType"`
	Destructors       []*xmlDestructor       `xml:"Destructor"`
	Enumerations      []*xmlEnumeration      `xml:"Enumeration"`
	Fields            []*xmlField            `xml:"Field"`
	Files             []*xmlFile             `xml:"File"`
	Functions         []*xmlFunction         `xml:"Function"`
	FunctionTypes     []*xmlFunctionType     `xml:"FunctionType"`
	FundamentalTypes  []*xmlFundamentalType  `xml:"FundamentalType"`
	Methods           []*xmlMethod           `xml:"Method"`
	MethodTypes       []*xmlMethodType       `xml:"MethodType"`
	Namespaces        []*xmlNamespace        `xml:"Namespace"`
	NamespaceAliases  []*xmlNamespaceAlias   `xml:"NamespaceAlias"`
	OperatorFunctions []*xmlOperatorFunction `xml:"OperatorFunction"`
	OperatorMethods   []*xmlOperatorMethod   `xml:"OperatorMethod"`
	OffsetTypes       []*xmlOffsetType       `xml:"OffsetType"`
	PointerTypes      []*xmlPointerType      `xml:"PointerType"`
	ReferenceTypes    []*xmlReferenceType    `xml:"ReferenceType"`
	Structs           []*xmlStruct           `xml:"Struct"`
	Typedefs          []*xmlTypedef          `xml:"Typedef"`
	Unimplementeds    []*xmlUnimplemented    `xml:"Unimplemented"`
	Unions            []*xmlUnion            `xml:"Union"`
	Variables         []*xmlVariable         `xml:"Variable"`
}

func (x *xmlTree) printStats() {

	fmt.Printf("loaded [%d] namespaces\n", len(x.Namespaces))
	fmt.Printf("loaded [%d] classes\n", len(x.Classes))
	fmt.Printf("loaded [%d] functions\n", len(x.Functions))
	fmt.Printf("loaded [%d] typedefs\n", len(x.Typedefs))
	fmt.Printf("loaded [%d] files\n", len(x.Files))
	fmt.Printf("loaded [%d] structs\n", len(x.Structs))
	fmt.Printf("loaded [%d] unions\n", len(x.Unions))
	fmt.Printf("loaded [%d] pointer-types\n", len(x.PointerTypes))
	fmt.Printf("loaded [%d] ref-types\n", len(x.ReferenceTypes))
	fmt.Printf("loaded [%d] cv-types\n", len(x.CvQualifiedTypes))
	fmt.Printf("loaded [%d] fundamentals\n", len(x.FundamentalTypes))

	if g_dbg {
		for i,_ := range x.FundamentalTypes {
			fmt.Printf(" => %v\n", x.FundamentalTypes[i])
		}
	}
}

//func (x *xmlTree) visit(node i_id) i_visitor {
//	return nil
//}

func (x *xmlTree) id() string {
	return "__0__"
}

// idDB associates the gccxml string id of a type to its parsed xmlFoobar struct
type idDB map[string]i_id

type xmlArgument struct {
	Attributes string `xml:"attributes,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	Default    string `xml:"default,attr"`
}

func (x *xmlArgument) id() string {
	return x.Id
}

type xmlArray struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"`
	Id         string `xml:"id,attr"`
	Max        string `xml:"max,attr"`
	Min        string `xml:"min,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlArray) id() string {
	return x.Id
}

func (x *xmlArray) name() string {
	t := g_ids[x.Type]
	tt := t.(i_name)
	return fmt.Sprintf("%s[%s]", tt.name(), x.Size)
}

type xmlBase struct {
	Type    string `xml:"type,attr"`
	Access  string `xml:"access,attr"`
	Virtual string `xml:"virtual,attr"`
	Offset  string `xml:"offset,attr"`
}

type xmlEllipsis struct {
	XMLName xml.Name `xml:"Ellipsis"`
}

type xml_record struct {
	Abstract   string `xml:"abstract,attr"`
	Access     string `xml:"access,attr"` // default "public"
	Align      string `xml:"align,attr"`
	Artificial string `xml:"artificial,attr"`
	Attributes string `xml:"attributes,attr"`
	XBases     string `xml:"bases,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Incomplete string `xml:"incomplete,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Members    string `xml:"members,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`

	Bases []xmlBase `xml:"Base"`
}

func (x *xml_record) id() string {
	return x.Id
}

func (x *xml_record) name() string {
	return x.Name
}

type xmlClass struct {
	xml_record
}

type xmlConstructor struct {
	Access     string `xml:"access,attr"`     // default "public"
	Artifical  string `xml:"artifical,attr"`  // implied
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Throw      string `xml:"throw,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlConstructor) id() string {
	return x.Id
}

type xmlConverter struct {
	Access     string `xml:"access,attr"`     // default "public"
	Attributes string `xml:"attributes,attr"` // implied
	Const      string `xml:"const,attr"`      // default "0"
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Returns    string `xml:"returns,attr"`
	Throw      string `xml:"throw,attr"`
	Virtual    string `xml:"virtual,attr"` // default "0"
}

func (x *xmlConverter) id() string {
	return x.Id
}

type xmlCvQualifiedType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Const      string `xml:"const,attr"`
	Id         string `xml:"id,attr"`
	Restrict   string `xml:"restrict,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
	Volatile   string `xml:"volatile,attr"`
}

func (x *xmlCvQualifiedType) id() string {
	return x.Id
}

func (x *xmlCvQualifiedType) name() string {
	t := g_ids[x.Type]
	return t.(i_name).name()
}

type xmlDestructor struct {
	Access     string `xml:"access,attr"`     // default "public"
	Artifical  string `xml:"artifical,attr"`  // implied
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Throw      string `xml:"throw,attr"`
	Virtual    string `xml:"virtual,attr"` // default "0"
}

func (x *xmlDestructor) id() string {
	return x.Id
}

type xmlEnumValue struct {
	Init string `xml:"init,attr"`
	Name string `xml:"name,attr"`
}

type xmlEnumeration struct {
	Access     string `xml:"access,attr"` // default "public"
	Align      string `xml:"align,attr"`
	Artifical  string `xml:"artifical,attr"`  // implied
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`

	EnumValues []xmlEnumValue `xml:"EnumValue"`
}

func (x *xmlEnumeration) id() string {
	return x.Id
}

func (x *xmlEnumeration) name() string {
	return x.Name
}

type xmlField struct {
	Access     string `xml:"access,attr"`     // default "public"
	Attributes string `xml:"attributes,attr"` // implied
	Bits       string `xml:"bits,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Mutable    string `xml:"mutable,attr"`
	Name       string `xml:"name,attr"`
	Offset     string `xml:"offset,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlField) id() string {
	return x.Id
}

func (x *xmlField) name() string {
	return x.Name
}

type xmlFile struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

func (x *xmlFile) id() string {
	return x.Id
}

func (x *xmlFile) name() string {
	return x.Name
}

type xmlFunction struct {
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Returns    string `xml:"returns,attr"`
	Throw      string `xml:"throw,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlFunction) id() string {
	return x.Id
}

func (x *xmlFunction) name() string {
	return x.Name
}

type xmlFunctionType struct {
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlFunctionType) id() string {
	return x.Id
}

type xmlFundamentalType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`
}

func (x *xmlFundamentalType) String() string {
	return fmt.Sprintf(
		`builtin{"%s", size=%s, align=%s}`,
		x.Name, x.Size, x.Align,
		)
}

func (x *xmlFundamentalType) id() string {
	return x.Id
}

func (x *xmlFundamentalType) name() string {
	return x.Name
}

type xmlMethod struct {
	Access      string `xml:"access,attr"`     // default "public"
	Attributes  string `xml:"attributes,attr"` // implied
	Const       string `xml:"const,attr"`
	Context     string `xml:"context,attr"`
	Demangled   string `xml:"demangled,attr"`
	Endline     string `xml:"endline,attr"`
	Extern      string `xml:"extern,attr"` // default "0"
	File        string `xml:"file,attr"`
	Id          string `xml:"id,attr"`
	Line        string `xml:"line,attr"`
	Location    string `xml:"location,attr"`
	Mangled     string `xml:"mangled,attr"`
	Name        string `xml:"name,attr"`
	PureVirtual string `xml:"pure_virtual,attr"` // default "0"
	Returns     string `xml:"returns,attr"`
	Static      string `xml:"static,attr"` // default "0"
	Throw       string `xml:"throw,attr"`
	Virtual     string `xml:"virtual,attr"` // default "0"

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlMethod) id() string {
	return x.Id
}

type xmlMethodType struct {
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Id         string `xml:"id,attr"`
	Returns    string `xml:"returns,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlMethodType) id() string {
	return x.Id
}

type xmlNamespace struct {
	Attributes string `xml:"attributes,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Id         string `xml:"id,attr"`
	Mangled    string `xml:"mangled,attr"`
	Members    string `xml:"members,attr"`
	Name       string `xml:"name,attr"`
}

func (x *xmlNamespace) id() string {
	return x.Id
}

type xmlNamespaceAlias struct {
	Context   string `xml:"context,attr"`
	Demangled string `xml:"demangled,attr"`
	Id        string `xml:"id,attr"`
	Mangled   string `xml:"mangled,attr"`
	Name      string `xml:"name,attr"`
	Namespace string `xml:"namespace,attr"`
}

func (x *xmlNamespaceAlias) id() string {
	return x.Id
}

type xmlOffsetType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	BaseType   string `xml:"basetype,attr"`
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlOffsetType) id() string {
	return x.Id
}

type xmlOperatorFunction struct {
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Endline    string `xml:"endline,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Returns    string `xml:"returns,attr"`
	Throw      string `xml:"throw,attr"`

	Arguments []xmlArgument `xml:"Argument"`
	Ellipsis  xmlEllipsis   `xml:"Ellipsis"`
}

func (x *xmlOperatorFunction) id() string {
	return x.Id
}

type xmlOperatorMethod struct {
	xmlMethod
}

type xmlPointerType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlPointerType) id() string {
	return x.Id
}

func (x *xmlPointerType) name() string {
	t := g_ids[x.Type]
	switch tt := t.(type) {
	case i_name:
		return tt.name()+"*"
	default:
	}
	return "<unnamed>*"
}

type xmlReferenceType struct {
	Align      string `xml:"align,attr"`
	Attributes string `xml:"attributes,attr"` // implied
	Id         string `xml:"id,attr"`
	Size       string `xml:"size,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlReferenceType) id() string {
	return x.Id
}

type xmlStruct struct {
	xml_record
}

type xmlTypedef struct {
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlTypedef) id() string {
	return x.Id
}

func (x *xmlTypedef) name() string {
	return x.Name
}

type xmlUnimplemented struct {
	Function     string `xml:"function,attr"`
	Id           string `xml:"id,attr"`
	Node         string `xml:"node,attr"`
	TreeCode     string `xml:"tree_code,attr"`
	TreeCodeName string `xml:"tree_code_name,attr"` // template_type_parm|typename_type|using_decl
}

func (x *xmlUnimplemented) id() string {
	return x.Id
}

type xmlUnion struct {
	Abstract   string `xml:"abstract,attr"`   // default "0"
	Access     string `xml:"access,attr"`     // default "public"
	Align      string `xml:"align,attr"`      // implied
	Artificial string `xml:"artificial,attr"` // "0"
	Attributes string `xml:"attributes,attr"` // implied
	Bases      string `xml:"bases,attr"`
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Incomplete string `xml:"incomplete,attr"` // default "0"
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Size       string `xml:"size,attr"`
}

func (x *xmlUnion) id() string {
	return x.Id
}

func (x *xmlUnion) name() string {
	return x.Name
}

type xmlVariable struct {
	Access     string `xml:"access,attr"`     // default "public"
	Artificial string `xml:"artificial,attr"` // "0"
	Attributes string `xml:"attributes,attr"` // implied
	Context    string `xml:"context,attr"`
	Demangled  string `xml:"demangled,attr"`
	Extern     string `xml:"extern,attr"` // default "0"
	File       string `xml:"file,attr"`
	Id         string `xml:"id,attr"`
	Init       string `xml:"init,attr"`
	Line       string `xml:"line,attr"`
	Location   string `xml:"location,attr"`
	Mangled    string `xml:"mangled,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
}

func (x *xmlVariable) id() string {
	return x.Id
}

func (x xmlNamespace) String() string {
	return fmt.Sprintf(`Namespace{%s}`, x.Name)
}

func (x xml_record) infos() string {
	bases := ""
	if len(x.Bases) > 0 {
		bases = ", bases=["
		for i, b := range x.Bases {
			if i > 0 {
				bases += " "
			}
			bases += b.Offset
			if i+1 != len(x.Bases) {
				bases += ", "
			}
		}
		bases += "]"
	}
	nmbrs := len(strings.Split(x.Members, " "))
	return fmt.Sprintf(`{%s, size=%s%s, nmbrs=%d}`,
		x.Name,
		x.Size,
		bases,
		nmbrs)
}

func (x xmlStruct) String() string {
	return "Struct" + x.xml_record.infos()
}

func (x xmlClass) String() string {
	return "Class" + x.xml_record.infos()
}

func (x xmlFunction) String() string {
	return fmt.Sprintf(`Function{%s, returns=%s, nargs=%d, ellipsis=%s}`,
		x.Name,
		x.Returns,
		len(x.Arguments),
		x.Ellipsis)
}


