package providers

import "fmt"

type Registry struct {
	byName  map[string]Provider
	byModel map[string][]Provider
}

func NewRegistry(list []Provider) *Registry {
	r := &Registry{
		byName:  make(map[string]Provider),
		byModel: make(map[string][]Provider),
	}
	for _, p := range list {
		r.byName[p.Name()] = p
		for _, m := range p.SupportedModels() {
			r.byModel[m] = append(r.byModel[m], p)
		}
	}
	return r
}

func (r *Registry) GetByName(name string) (Provider, error) {
	p, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return p, nil
}

func (r *Registry) GetByModel(model string) ([]Provider, error) {
	ps, ok := r.byModel[model]
	if !ok || len(ps) == 0 {
		return nil, fmt.Errorf("no provider found for model %q", model)
	}
	return ps, nil
}
