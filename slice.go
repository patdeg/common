package common

// Add the string 'element' if not already in the array 'list'. Return the new array
func AddIfNotExists(element string, list []string) []string {
	for _, item := range list {
		if item == element {
			return list
		}
	}
	return append(list, element)
}

func AddIfNotExistsGeneric(element interface{}, list []interface{}) []interface{} {
	for _, item := range list {
		if item == element {
			return list
		}
	}
	return append(list, element)
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
