


main.go: main-lit.bt
	lit -f main-lit.bt -o main.go > main.go && goimports -w main.go && cat main.go

exp1.go: 
	lit -f main-lit.bt -o exp1.go > /tmp/a.go && goimports -w /tmp/a.go && cat /tmp/a.go


exp2.go: 
	lit -f main-lit.bt -o exp2.go > /tmp/a.go && goimports -w /tmp/a.go && cat /tmp/a.go

a.html:
	barkdown < ./main-lit.bt > /tmp/a.html && open /tmp/a.html 
