function nodeToHtml(node) {
  const s =  _nodeToHtml(node);
  // F:
  // This makes paragraphs from
  // text separated by blank lines.
  return paragraphs(s);
}

function _nodeToHtml(node) {
  if (node instanceof Blob) {
    return node.text
  }

  if (node instanceof Code) {
    // F:
    // Without this inline code appers on separate line.
    if(node.tag === "`") {
      return "<code>" + node.text + "</code>"
    }

    return "<pre>" + node.text + "</pre>";
  }

  if (node instanceof Bag) {

    // F:
    // This is why you don't see
    // LOOK HERE in the preview.
    if(node.tag.startsWith("-")) {
      return "";
    }

    // F:
    // This makes lists work.
    if(node.tag === "ol") {
      return makeList(node);
    }

    // F:
    // This makes table.
    if(node.tag === "table") {
      return makeTable(node);
    }

    let s = "";
    for (const k of node.kids) {
      s += _nodeToHtml(k);
    }
    if (!node.tag) return s;
    return "<" + node.tag + ">" + s + "</" + node.tag + ">";
  }

  return "";
}

function paragraphs(text) {
  const ps = text.split("\n\n");
  for(const [i, p] of ps.entries()) {
    if(p.trim() !== "") {
      ps[i] = "<p>" + p + "</p>";
    }
  }
  return ps.join("");
}

function makeTable(bg) {
  const lines = textOf(bg)
    .split("\n")
    .filter(x => x.trim() !== "");

  function td(s) {
    return '<td>'
      + s.split('&').join('</td><td>')
      + '</td>';
  }

  for(const [i, line] of lines.entries()) {
    if(line.trim() !== "") {
      lines[i] =  '<tr>' + td(line) + "</tr>";
    }
  }

  return  '<table border="1" style="border-collapse: collapse;">'
  + lines.join("\n") + "</table>"
}

function makeList(bg) {
  const lines = textOf(bg)
    .split("\n")
    .filter(x => x.trim() !== "");

  const otag = `<${bg.tag}>`;
  const ctag = `</${bg.tag}>`;

  const strs = [];
  const stack = [];

  function prefixLen(s) {
    return s.length - s.trimStart().length;
  }

  function push(a) {
    stack.push(a);
    strs.push(otag);
  }

  function pop() {
    stack.pop();
    strs.push(ctag);
  }

  function top() {
    return stack[stack.length - 1];
  }

  push(0);

  for (const line of lines) {
    const l = prefixLen(line);
    if (l > top()) {
      push(l);
    } else if (l < top()) {
      pop();
    }

    const content = line.trimStart().slice(2);
    const h = toHtml(`#{${content}}`);

    strs.push(`<li>${h}</li>`);
  }

  while (stack.length > 0) {
    pop();
  }

  return strs.slice(1, strs.length - 1).join("\n");
}

function textOf(x) {
  let s = "";
  for (const k of x.kids) {
    s += k.string();
  }
  return s;
}