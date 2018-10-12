package diff

import (
	"log"

	"github.com/milesbxf/kryp/pkg/k8s"

	_ "k8s.io/kubernetes/pkg/master"
)

func GetFileDiff(filename string, helper *k8s.ResourceHelper) ([]Diff, error) {
	resources, err := helper.NewResourcesFromFilename(filename)
	if err != nil {
		log.Printf("Error getting resource: %v", err)
		return []Diff{EmptyDiff{}}, err
	}
	diffs := []Diff{}
	for _, r := range resources {
		d, err := processResource(r, helper)
		if err != nil {
			return diffs, err
		}
		diffs = append(diffs, d)
	}
	return diffs, nil
}
func processResource(resource *k8s.Resource, helper *k8s.ResourceHelper) (Diff, error) {
	log.Printf("Setting defaults for object %v", resource.Object)
	defaultedObj := k8s.GetWithDefaults(resource.Object)
	meta := DiffMeta{Resource: resource}

	serverObj, err := resource.Get()
	if k8s.IsNotFoundError(err) {
		return NotPresentOnServerDiff{DiffMeta: meta}, nil
	}
	deltas, err := calculateDiff(defaultedObj, serverObj)
	if err != nil {
		log.Printf("Error calculating deltas: %v", err)
		return EmptyDiff{}, err
	}
	log.Printf("Found %d deltas", len(deltas))
	if len(deltas) == 0 {
		return EmptyDiff{DiffMeta: meta}, nil
	}

	filtered := deltas
	for _, f := range []DeltaFilter{MetadataFilter} {
		filtered = f(filtered)
	}
	return ChangesPresentDiff{DiffMeta: meta, deltas: filtered}, nil
}

var empty = struct{}{}

func (d EmptyDiff) Pretty() string   { return "" }
func (ed EmptyDiff) Deltas() []Delta { return []Delta{} }

func (d ChangesPresentDiff) Deltas() []Delta { return d.deltas }

func (d NotPresentOnServerDiff) Pretty() string  { return "" }
func (d NotPresentOnServerDiff) Deltas() []Delta { return []Delta{} }