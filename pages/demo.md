<nav>
This is a navigation block.
</nav>

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

*Italic*

**Bold**

~~Strikethrough~~

[Link](http://a.com)

[[miniwiki|Internal link]]

![Image](https://commonmark.org/help/images/favicon.png)

> Blockquote

* List
* List
* List

1. One
2. Two
3. Three

Horizontal rule:

---

`Inline code` with backticks

```sh
# code block
print '3 backticks or'
print 'indent 4 spaces'
```

www.commonmark.org

- [ ] foo
- [x] bar

| foo | bar |
| --- | --- |
| baz | bim |

That's some text with a footnote.[^1]

[^1]: And that's the footnote.

| Option | Description |
| ------ | ----------- |
| data   | path to data files to supply the data that will be passed into templates. |
| engine | engine to be used for processing templates. Handlebars is the default. |
| ext    | extension to be used for dest files. |

Right aligned columns

| Option | Description |
| ------:| -----------:|
| data   | path to data files to supply the data that will be passed into templates. |
| engine | engine to be used for processing templates. Handlebars is the default. |
| ext    | extension to be used for dest files. |


![Alt text][id]

With a reference later in the document defining the URL location:

[id]: https://octodex.github.com/images/dojocat.jpg  "The Dojocat"
