package inference

import "github.com/scaleway/scaleway-sdk-go/api/inference/v1"

func flattenNodeSupport(nodesSupportInfo []*inference.ModelSupportInfo) []any {
	if len(nodesSupportInfo) == 0 {
		return nil
	}

	var result []any

	for _, nodeSupport := range nodesSupportInfo {
		if nodeSupport == nil {
			continue
		}

		for _, node := range nodeSupport.Nodes {
			flattenQuantization := make([]any, 0, len(node.Quantizations))

			for _, quantization := range node.Quantizations {
				if quantization == nil {
					continue
				}

				flattenQuantization = append(flattenQuantization, map[string]any{
					"quantization_bits": quantization.QuantizationBits,
					"allowed":           quantization.Allowed,
					"max_context_size":  quantization.MaxContextSize,
				})
			}

			result = append(result, map[string]any{
				"node_type_name": node.NodeTypeName,
				"quantization":   flattenQuantization,
			})
		}
	}

	return result
}
