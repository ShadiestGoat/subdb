IFS=""

tpl_out=$(cat "tpl/$1.yaml" | tpl -f "tpl/$1.tpl" -d yaml -n "$1.tpl")

echo $tpl_out | gofmt -s > "$1.go" || echo $tpl_out