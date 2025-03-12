package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func compareSelectors(a, b []metav1.LabelSelectorRequirement) (bool, error) {
	aLabelSelector := metav1.LabelSelector{
		MatchExpressions: a,
	}
	bLabelSelector := metav1.LabelSelector{
		MatchExpressions: b,
	}

	selectorA, err := metav1.LabelSelectorAsSelector(&aLabelSelector)
	if err != nil {
		return false, err
	}

	selectorB, err := metav1.LabelSelectorAsSelector(&bLabelSelector)
	if err != nil {
		return false, err
	}

	return selectorA.String() == selectorB.String(), nil
}

func getNamespacedNameFromObject(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}
