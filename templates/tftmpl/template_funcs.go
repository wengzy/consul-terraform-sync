package tftmpl

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/consul/api"
)

// HCLTmplFuncMap are template functions for rendering HCL
var HCLTmplFuncMap = map[string]interface{}{
	// "hclServiceHealth": hcat.HCLServiceHealth,
	"hclString":             HCLString,
	"hclStringList":         HCLStringList,
	"hclStringMap":          HCLStringMap,
	"hclConsulHealthChecks": HCLConsulHealthChecks,
}

// func HCLServiceHealth(s *serviceHealth) string {
// 	f := hclwrite.NewEmptyFile()
// 	gohcl.EncodeIntoBody(s, f.Body())

// 	var buf bytes.Buffer
// 	f.WriteTo(&buf)
// 	return buf.String()
// }

func HCLConsulHealthChecks(checks api.HealthChecks, indent int) string {
	if len(checks) == 0 {
		return "[]"
	}
	checksList := make([]string, len(checks))
	for i, s := range checks {
		checksList[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("[%s]", strings.Join(checksList, ", "))

	return fmt.Sprint(checks)
}

func hclConsulHealthCheck(check api.HealthCheck) string {
	// for each attribute and nested objects, convert to HCL attr name and type
	// there are no tags so we'll have to either dynamically reflect to get the type
	// or hard code the health check attributes

	attrs := []struct {
		name   string
		format string
		value  interface{}
	}{
		{
			"node",
			"%s = \"%s\"",
			check.Node,
		}, {
			"check_id",
			"%s = \"%s\"",
			check.CheckID,
		}, {
			"name",
			"%s = \"%s\"",
			check.Name,
		}, {
			"status",
			"%s = \"%s\"",
			check.Status,
			// TODO notes, output, service_id, service_name, type, namespace
		}, {
			"service_tags",
			"%s = {{HCLStringList .ServiceTags}}",
			check.ServiceTags,
		}, {
			"definition",
			fmt.Sprint("%s = ", hclHealthCheckDefintion(check.Definition)),
			check.Definition,
		},
	}

	var hclRaw string
	for _, attr := range attrs {
		hclRaw += fmt.Sprintf(attr.format, attr.name, attr.value)
	}
	return fmt.Sprintf("{%s}", hclRaw)
}

func hclHealthCheckDefintion(def api.HealthCheckDefinition) string {
	attrs := []struct {
		name   string
		format string
		value  interface{}
	}{
		{
			"http",
			"%s = \"%s\"",
			def.HTTP,
		}, {
			name: "header",
			// some complex typing to convert map[string][]string into HCL
		}, {
			"method",
			"%s = \"%s\"",
			def.Method,
		}, {
			"body",
			"%s = \"%s\"",
			def.Body,
		}, {
			"tls_skip_verify",
			"%s = %s",
			def.TLSSkipVerify,
		},
	}

	var hclRaw string
	for _, attr := range attrs {
		hclRaw += fmt.Sprintf(attr.format, attr.name, attr.value)
	}
	return fmt.Sprintf("{%s}", hclRaw)
}

// HCLString formats a string into HCL string with null as the default
func HCLString(s string) string {
	if s == "" {
		return "null"
	}
	return fmt.Sprintf("%q", s)
}

// HCLStringList formats a list of strings into HCL
func HCLStringList(l []string) string {
	if len(l) == 0 {
		return "[]"
	}

	hclList := make([]string, len(l))
	for i, s := range l {
		hclList[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("[%s]", strings.Join(hclList, ", "))
}

// HCLStringMap formats a map of strings into HCL
func HCLStringMap(m map[string]string, indent int) string {
	if len(m) == 0 {
		return "{}"
	}

	sortedKeys := sortKeys(m)

	if indent < 1 {
		keyValues := make([]string, len(m))
		for i, k := range sortedKeys {
			v := m[k]
			keyValues[i] = fmt.Sprintf("%s = \"%s\"", k, v)
		}
		return fmt.Sprintf("{ %s }", strings.Join(keyValues, ", "))
	}

	// Find the longest key to align values with proper Terraform fmt spacing
	var longestKeyLen int
	for _, k := range sortedKeys {
		keyLen := len(k)
		if longestKeyLen < keyLen {
			longestKeyLen = keyLen
		}
	}

	indentStr := strings.Repeat("  ", indent)
	indentStrClosure := strings.Repeat("  ", indent-1)

	var keyValues string
	for _, k := range sortedKeys {
		v := m[k]
		tfFmtSpaces := strings.Repeat(" ", longestKeyLen-len(k))
		keyValues = fmt.Sprintf("%s\n%s%s%s = \"%s\"", keyValues, indentStr, k, tfFmtSpaces, v)
	}
	return fmt.Sprintf("{%s\n%s}", keyValues, indentStrClosure)
}

func sortKeys(m map[string]string) []string {
	sorted := make([]string, 0, len(m))
	for key := range m {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)
	return sorted
}
