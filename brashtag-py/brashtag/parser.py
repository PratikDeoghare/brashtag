from __future__ import annotations
import sys
from dataclasses import dataclass

@dataclass
class Bag:
	tag: str
	kids: list[Bag | Blob | Code]

	def __str__(self):
		s = ''.join(str(k) for k in self.kids)
		return f'#{self.tag}{{{s}}}'


@dataclass
class Blob:
	text: str

	def __str__(self):
		return str(self.text)


@dataclass
class Code:
	tag: str
	text: str

	def __str__(self):
		return f"{self.tag}{self.text}{self.tag}"

@dataclass
class Error:
	s: str

	def __str__(self):
		return self.s

def parse_bag(s: str) -> tuple[str, Bag|Error]:
	if len(s) == 0:
		return s, Error('empty input')
	if s[0] != '#':
		return s, Error('expected #')
	i = 0
	while i < len(s) and s[i] != '{':
		i += 1

	if i >= len(s) or s[i] != '{':
		return s, Error('unopened bag')

	root = Bag(s[1:i], [])

	s = s[i+1:]
	while len(s) > 0:
		x : Bag | Blob | Code | Error
		if s[0] == '#':
			s, x = parse_bag(s)
			if isinstance(x, Error):
				return s, x
			root.kids.append(x)
		elif s[0] == '`':
			s, x = parse_code(s)
			if isinstance(x, Error):
				return s, x
			root.kids.append(x)
		elif s[0] == '}':
			s = s[1:]
			return s, root
		else:
			s, x = parse_blob(s)
			if isinstance(x, Error):
				return s, x
			root.kids.append(x)


	return s, Error('bag not closed')

def parse_blob(s: str) -> tuple[str, Blob|Error]:
	if len(s) == 0:
		return s, Error('empty input')
	i = 0
	while i < len(s) and s[i] != '#' and s[i] != '`' and s[i] != '}':
		i += 1
	if i == 0:
		return s, Error('empty blob')
	rest = s[i:]
	blob = Blob(s[:i])
	return rest, blob


def parse_code(s: str) -> tuple[str, Code|Error]:
	if len(s) == 0:
		return s, Error('empty input')
	i = 0
	while i < len(s) and s[i] == '`':
		i+=1
	tag = s[:i]
	k = len(tag)
	codeStart = i
	while i < len(s) - k and s[i:i+k] != tag:
		i += 1
	if s[i:i+k] == tag:
		return s[i+k:], Code(tag, s[codeStart:i])
	return s[codeStart:], Error(f'no end tag "{tag}"')

def parse(text :str) -> Bag|Error:
	rest, x = parse_bag(str(text))
	if isinstance(x, Error):
		n = min(100, len(rest))
		return Error(f'{x}: {rest[:n]}')
	return x

def run_tests():
	def run(f, text, expected):
		assert f(text) == expected, f(text)

	run(
		parse_bag,
		'#foo{bar}',
		('', Bag(tag='foo', kids=[Blob(text='bar')])),
	)

	run(
		parse_blob,
		'hello world!!`code`',
		('`code`', Blob(text='hello world!!')),
	)

	run(
		parse_blob,
		'#{This is not just a blob.}',
		('#{This is not just a blob.}', Error('empty blob')),
	)


	text = '''```
		Code looks like `code`.
	```the rest of it'''
	run(
		parse_code,
		text,
		('the rest of it', Code(tag='```', text='''
		Code looks like `code`.
	''')),
	)


	run(
		parse_code,
		'``This is unterminated code. Should error.`',
		('This is unterminated code. Should error.`', Error('no end tag "``"')),
	)

	run(
		parse_bag,
		'#my tag {is in my bag #of{another bag `code`} and the code is ``123``.}',
		('',
			Bag(tag="my tag ",
				kids = [
					Blob(text="is in my bag "),
					Bag(tag="of",
						kids = [
							Blob(text="another bag "),
							Code(tag="`", text="code"),
						],
					),
					Blob(text=" and the code is "),
					Code(tag="``", text="123"),
					Blob(text=".")
				],
			)
		),
	)


if __name__ == '__main__':
	if sys.argv[1] == 'test':
		run_tests()
	elif sys.argv[1] == 'parse':
		print(repr(parse(sys.stdin.read())))

