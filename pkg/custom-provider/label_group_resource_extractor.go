package provider

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"

	pmodel "github.com/prometheus/common/model"
)

// labelGroupResExtractor extracts schema.GroupResources from series labels.
type labelGroupResExtractor struct {
	regex *regexp.Regexp

	resourceInd int
	groupInd    *int
	mapper      apimeta.RESTMapper
}

// newLabelGroupResExtractor creates a new labelGroupResExtractor for labels whose form
// matches the given template.  It does so by creating a regular expression from the template,
// so anything in the template which limits resource or group name length will cause issues.
func newLabelGroupResExtractor(labelTemplate *template.Template) (*labelGroupResExtractor, error) {
	labelRegexBuff := new(bytes.Buffer)
	if err := labelTemplate.Execute(labelRegexBuff, schema.GroupResource{"(?P<group>.+?)", "(?P<resource>.+?)"}); err != nil {
		return nil, fmt.Errorf("unable to convert label template to matcher: %v", err)
	}
	if labelRegexBuff.Len() == 0 {
		return nil, fmt.Errorf("unable to convert label template to matcher: empty template")
	}
	labelRegexRaw := "^" + labelRegexBuff.String() + "$"
	labelRegex, err := regexp.Compile(labelRegexRaw)
	if err != nil {
		return nil, fmt.Errorf("unable to convert label template to matcher: %v", err)
	}

	var groupInd *int
	var resInd *int

	for i, name := range labelRegex.SubexpNames() {
		switch name {
		case "group":
			ind := i // copy to avoid iteration variable reference
			groupInd = &ind
		case "resource":
			ind := i // copy to avoid iteration variable reference
			resInd = &ind
		}
	}

	if resInd == nil {
		return nil, fmt.Errorf("must include at least `{{.Resource}}` in the label template")
	}

	return &labelGroupResExtractor{
		regex:       labelRegex,
		resourceInd: *resInd,
		groupInd:    groupInd,
	}, nil
}

// GroupResourceForLabel extracts a schema.GroupResource from the given label, if possible.
// The second argument indicates whether or not a potential group-resource was found in this label.
func (e *labelGroupResExtractor) GroupResourceForLabel(lbl pmodel.LabelName) (schema.GroupResource, bool) {
	matchGroups := e.regex.FindStringSubmatch(string(lbl))
	if matchGroups != nil {
		group := ""
		if e.groupInd != nil {
			group = matchGroups[*e.groupInd]
		}

		return schema.GroupResource{
			Group:    group,
			Resource: matchGroups[e.resourceInd],
		}, true
	}

	return schema.GroupResource{}, false
}
