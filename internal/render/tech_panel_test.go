package render

import "testing"

func TestTechConnectorMidYSpreadsMultiRequirementEdges(t *testing.T) {
	parent := techNode{x: 100, y: 100}
	child := techNode{x: 220, y: 220}

	first := techConnectorMidY(parent, child, 0, 2)
	second := techConnectorMidY(parent, child, 1, 2)
	if !(first < second) {
		t.Fatalf("multi prerequisite connector lanes should spread, got first=%.1f second=%.1f", first, second)
	}
}

func TestTechConnectorMidYStaysBetweenNodes(t *testing.T) {
	parent := techNode{x: 100, y: 100}
	child := techNode{x: 100, y: 220}

	midY := techConnectorMidY(parent, child, 3, 5)
	minY := parent.y + techNodeHeight/2 + techConnectorStem
	maxY := child.y - techNodeHeight/2 - techConnectorStem
	if midY < minY || midY > maxY {
		t.Fatalf("connector lane escaped node gap, got %.1f outside [%.1f, %.1f]", midY, minY, maxY)
	}
}
