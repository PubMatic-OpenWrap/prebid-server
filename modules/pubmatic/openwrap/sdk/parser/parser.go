package sdkparser

func ParseTemplateAndSetValues(template, source, target map[string]any) {
	for key, value := range template {
		srcValue, ok := source[key]
		if !ok {
			continue
		}

		switch tmplValue := value.(type) {
		case map[string]any:
			sourceMap, ok := srcValue.(map[string]any)
			if !ok {
				continue
			}

			targetMap, ok := target[key].(map[string]any)
			if !ok {
				targetMap = make(map[string]any)
				target[key] = targetMap
			}

			ParseTemplateAndSetValues(tmplValue, sourceMap, targetMap)
		case []any:
			sourceArray, ok := srcValue.([]any)
			if !ok {
				continue
			}

			// Check what happens if targetArray is not empty
			targetArray, ok := target[key].([]any)
			if !ok || targetArray == nil {
				targetArray = make([]any, len(sourceArray))
				target[key] = targetArray
			}

			for i := range tmplValue {
				if i >= len(sourceArray) {
					break
				}

				switch elem := tmplValue[i].(type) {
				case map[string]any:
					sourceMap, ok := sourceArray[i].(map[string]any)
					if !ok {
						continue
					}

					targetMap, ok := targetArray[i].(map[string]any)
					if !ok {
						targetMap = make(map[string]any)
						targetArray[i] = targetMap
					}
					ParseTemplateAndSetValues(elem, sourceMap, targetMap)
				default:
					target[key] = sourceArray
				}
			}
		default:
			switch tmplValue {
			case "string", "int", "float", "bool", "array", "object":
				target[key] = srcValue
			default:
				target[key] = tmplValue
			}
		}
	}
}
