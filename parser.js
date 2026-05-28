
export class Bag {
  constructor(tag, kids = []) {
    this.tag = tag;
    this.kids = kids;
  }

  addKids(...kids) {
    this.kids.push(...kids);
  }

  setTag(tag) {
    this.tag = tag;
  }

  string() {
    return `#${this.tag}{${this.kids.map(k => k.string()).join("")}}`;
  }
}

export class Code {
  constructor(tag, text) {
    this.tag = tag;
    this.text = text;
  }

  setTag(tag) {
    this.tag = tag;
  }

  setText(text) {
    this.text = text;
  }

  string() {
    return `${this.tag}${this.text}${this.tag}`;
  }
}

export class Blob {
  constructor(text) {
    this.text = text;
  }

  setText(text) {
    this.text = text;
  }

  string() {
    return this.text;
  }
}

export function parse(text) {

  const result = parseBag(text);

  if (result.error) {
    throw new Error(result.error);
  }

  if (result.rest.length !== 0) {
    throw new Error(
      "unexpected trailing input: " +
      result.rest.slice(0, 100)
    );
  }

  return result.node;
}

export function parseBag(text) {

  const root = new Bag("");

  let c = 0;
  let bagOpen = false;

  while (c < text.length) {

    if (text[c] !== "{") {
      c++;
    } else {
      bagOpen = true;
      break;
    }
  }

  if (!bagOpen) {
    return {
      error: "unopened bag missing { after #",
      rest: text
    };
  }

  root.setTag(text.slice(1, c));

  text = text.slice(c + 1);

  while (text.length > 0) {

    if (text[0] === "#") {

      const result = parseBag(text);

      if (result.error) return result;

      root.addKids(result.node);

      text = result.rest;

      continue;
    }

    if (text[0] === "`") {

      const result = parseCode(text);

      if (result.error) return result;

      root.addKids(result.node);

      text = result.rest;

      continue;
    }

    if (text[0] === "}") {

      return {
        node: root,
        rest: text.slice(1)
      };
    }

    const result = parseBlob(text);

    if (result.error) return result;

    root.addKids(result.node);

    text = result.rest;
  }

  return {
    error: "bag not closed",
    rest: text
  };
}

export function parseCode(text) {

  const root = new Code("", "");

  let c = 0;

  while (c < text.length && text[c] === "`") {
    c++;
  }

  root.setTag(text.slice(0, c));

  let j = c;

  while (j < text.length - c + 1) {

    if (text.slice(j, j + c) !== root.tag) {
      j++;
    } else {

      root.setText(text.slice(c, j));

      return {
        node: root,
        rest: text.slice(j + c)
      };
    }
  }

  return {
    error: "code not closed",
    rest: text
  };
}

export function parseBlob(text) {

  const root = new Blob("");

  let j = 0;

  while (j < text.length) {

    switch (text[j]) {

      case "`":
      case "#":
      case "}":

        root.setText(text.slice(0, j));

        return {
          node: root,
          rest: text.slice(j)
        };

      case "{":

        return {
          error: "unexpected {"
        };

      default:
        j++;
    }
  }

  root.setText(text.slice(0, j));

  return {
    node: root,
    rest: text.slice(j)
  };
}

