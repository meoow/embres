package main

import (
	"code.google.com/p/go.net/html"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type foundnode struct {
	node *html.Node
	path string
}

func embed_file(pathOfFile string, w io.Writer) {
	fh, err := os.Open(pathOfFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	root, err := html.Parse(fh)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(root.FirstChild.NextSibling)
	founds := make([]*foundnode, 0, 10)
	dir := filepath.Dir(pathOfFile)
	find_nodes(root, &founds, dir)
	// for _, f := range founds {
	// 	fmt.Println("%#v\n", f)
	// }

	for _, n := range founds {
		embed_html(n)
	}
	html.Render(w, root)
}

func embed_html(fn *foundnode) {
	// fmt.Println(1, fn.path)
	fh, err := os.Open(fn.path)
	if err != nil {
		return
	}
	defer fh.Close()
	filebytes, err := ioutil.ReadAll(fh)
	// fmt.Println(2, fn.path)

	switch fn.node.Data {
	case "style", "script":
		new_node := new(html.Node)
		new_node.Type = html.TextNode
		new_node.Data = fmt.Sprintf("\n<!--\n%s\n-->\n", filebytes)
		fn.node.AppendChild(new_node)
		attr_map := attr_to_map(fn.node.Attr)
		fn.node.Attr = make([]html.Attribute, 0, len(attr_map)+1)
		found_type := false
		for idx, attr := range fn.node.Attr {
			if attr.Key == "src" {
				fn.node.Attr = append(fn.node.Attr[:idx], fn.node.Attr[idx+1:]...)
				break
			} else if attr.Key == "type" {
				found_type = true
			}
		}
		if !found_type {
			var text_type string
			if fn.node.Data == "style" {
				text_type = "text/css"
			} else {
				text_type = "text/javascript"
			}
			fn.node.Attr = append(fn.node.Attr, html.Attribute{Key: "type", Val: text_type})
		}
	case "link":
		style_node := new(html.Node)
		text_node := new(html.Node)
		style_node.Type = html.ElementNode
		style_node.Data = "style"
		style_node.Attr = []html.Attribute{{Key: "type", Val: "text/css"}}
		text_node.Type = html.TextNode
		text_node.Data = fmt.Sprintf("\n<!--\n%s\n-->\n", filebytes)
		style_node.AppendChild(text_node)
		if fn.node.Parent != nil {
			fn.node.Parent.InsertBefore(style_node, fn.node)
		}
		fn.node.Parent.RemoveChild(fn.node)
	case "img":
		base64img := base64.StdEncoding.EncodeToString(filebytes)
		for idx, attr := range fn.node.Attr {
			if attr.Key == "src" {
				imgtype := strings.TrimPrefix(filepath.Ext(attr.Val), ".")
				fn.node.Attr = append(fn.node.Attr[:idx], fn.node.Attr[idx+1:]...)
				fn.node.Attr = append(fn.node.Attr, html.Attribute{Key: "src", Val: fmt.Sprintf("data:image/%s;base64,%s", imgtype, base64img)})
				break
			}
		}
	}
}

func find_nodes(node *html.Node, founds *[]*foundnode, prefix string) {

	var attrs map[string]string

	for n := node.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode {
			continue
		}

		switch n.Data {
		case "link", "style", "script", "img":
			// fmt.Println(n.Data)
			attrs = attr_to_map(n.Attr)
		default:
			goto FIND_NEXT
		}

		switch n.Data {
		case "link":
			if _, ok := attrs["href"]; !ok {
				goto FIND_NEXT
			}
			if attrs["rel"] != "stylesheet" && attrs["type"] != "text/css" {
				goto FIND_NEXT
			}

			path, err := url.QueryUnescape(attrs["href"])
			if err != nil {
				path = attrs["href"]
			}
			path = filepath.Join(prefix, path)
			*founds = append(*founds, &foundnode{n, path})
		case "style":
			if attrs["type"] != "text/css" {
				goto FIND_NEXT
			}
			if _, ok := attrs["src"]; !ok {
				goto FIND_NEXT
			}
			path, err := url.QueryUnescape(attrs["src"])
			if err != nil {
				path = attrs["src"]
			}
			path = filepath.Join(prefix, path)
			*founds = append(*founds, &foundnode{n, path})
		case "script", "img":
			if _, ok := attrs["src"]; !ok {
				goto FIND_NEXT
			}
			path, err := url.QueryUnescape(attrs["src"])
			if err != nil {
				path = attrs["src"]
			}
			path = filepath.Join(prefix, path)
			*founds = append(*founds, &foundnode{n, path})
		}
	FIND_NEXT:
		find_nodes(n, founds, prefix)
	}
}

func attr_to_map(attrs []html.Attribute) map[string]string {
	attrmap := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		attrmap[attr.Key] = attr.Val
	}
	return attrmap
}
