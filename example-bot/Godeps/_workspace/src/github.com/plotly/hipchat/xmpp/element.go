package xmpp

import (
	"encoding/xml"
	"fmt"
)

type Element struct {
	StartElement xml.StartElement
	EndElement   xml.EndElement
	CharData     string
	Comment      xml.Comment
	ProcInst     xml.ProcInst
	Directive    xml.Directive
	Children     []*Element
	Parent       *Element
}

func NewElement(parent *Element) *Element {
	elem := new(Element)
	elem.Children = make([]*Element, 1)
	if parent != nil {
		parent.Children = append(parent.Children, elem)
		elem.Parent = parent
	}
	return elem
}

func (elem *Element) Name() xml.Name {
	return elem.StartElement.Name
}

func (elem *Element) NameWithSpace() string {
	return elem.Name().Local + elem.Name().Space
}

func (elem *Element) AttrMap() map[string]string {
	return ToMap(elem.StartElement.Attr)
}

func (elem *Element) IsMessage() bool {
	if elem.NameWithSpace() != "message"+NsJabberClient {
		return false
	}
	attr := elem.AttrMap()
	return attr["type"] == "groupchat" || attr["type"] == "chat"
}

type elemMatcher func(*Element) bool

func (elem *Element) FindChild(matcher elemMatcher) *Element {
	for _, child := range elem.Children {
		if child == nil {
			continue
		}
		if matcher(child) {
			return child
		}
	}
	return nil
}

func (elem *Element) String() string {
	result := fmt.Sprintf("(start:%v", elem.StartElement)
	charData := elem.CharData
	if charData != "" {
		result += fmt.Sprintf(", charData:%v", charData)
	}
	if len(elem.Children) > 0 {
		result += ", children:("
		for _, child := range elem.Children {
			if child == nil {
				continue
			}
			result += child.String()
		}
		result += ")"
	}
	result += ")"
	return result
}
