document.querySelectorAll('a').forEach(function (node) {
    var cloneLink = node.cloneNode(true)
    var sourcePath = node.href.replace(embedReg, "").replace(location.origin + location.pathname, '')
    var filename = node.href.replace(embedReg, "").replace(/(.*)?\//,'')
    cloneLink.innerText += ": " + filename
    var embedReg = /\?embed$/g
    if (!embedReg.test(node.href)) {
        return
    }
    var path = node.href.replace(embedReg, "")
    var onlineHref = GithubRepoURL + "/blob/main/" + sourcePath.replace(embedReg, "")
    var source = ""
    var text = fetch(path).then(function (res){
        if (res.status == 200) {
            return res.text()
        }
        return new Promise(function (resolve){
            resolve(null)
        })
    }).then(function (source){
        if (!source) {
            return
        }
        text = "// " + onlineHref + "\n" + source
        html = hljs.highlightAuto(text).value

        var box = document.createElement('div')
        box.style.position = "relative"
        var pre = document.createElement('pre')
        var code = document.createElement('code')

        code.innerHTML = html
        pre.appendChild(code)
        var sourceLink = document.createElement("a")
        sourceLink.innerText = sourcePath.replace(embedReg, "")
        sourceLink.href = onlineHref
        box.appendChild(sourceLink)
        box.appendChild(pre)
        var cloneSourceLink = sourceLink.cloneNode(true)
        cloneSourceLink.style.position = "relative"
        cloneSourceLink.style.top = "-1em"
        box.appendChild(cloneSourceLink)
        node.parentNode.replaceChild(box,node)
    })
})