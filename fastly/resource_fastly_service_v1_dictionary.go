package fastly

import (
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

type Dictionary struct {
	dictionary *gofastly.Dictionary
	items      []*gofastly.DictionaryItem
}

func flattenDictionaries(dictionariesList []*Dictionary) []map[string]interface{} {
	var dl []map[string]interface{}

	for _, d := range dictionariesList {
		nd := map[string]interface{}{
			"name": d.dictionary.Name,
			"id":   d.dictionary.ID,
		}

		if len(d.items) > 0 {
			is := make(map[string]string)
			for _, item := range d.items {
				is[item.ItemKey] = item.ItemValue
			}

			nd["items"] = is
		}
		dl = append(dl, nd)
	}
	return dl
}
