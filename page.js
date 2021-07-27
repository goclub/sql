!(function () {
    function markrunSidebar(settings) {
        settings = settings || {}
        settings.content = settings.content || document.body
        var map = {}
        var sidebar = document.createElement('ul')
        var childNodes = settings.content.querySelectorAll('h2,h3')
        var childNodesLength = childNodes.length
        var currentSubHeaderId = null
        childNodes.forEach(function(element){
            var id = element.id
            var text = element.innerText
            var data = {
                id: id,
                text: text
            }
            switch (element.tagName) {
                case 'H2':
                    currentSubHeaderId = id
                    map[id] = map[id] || data
                    break
                case 'H3':
                    if (currentSubHeaderId !== null) {
                        map[currentSubHeaderId].child = map[currentSubHeaderId].child ||[]
                        map[currentSubHeaderId].child.push(data)
                    }
                    break
                default:
            }
        })
        for(var key in map) {
            var item = map[key]
            var li = document.createElement('li')
            var link = document.createElement('a')
            link.innerHTML = item.text
            link.setAttribute('href', '#' + item.id)
            link.setAttribute('class', 'markdown-sidebar-link')
            li.appendChild(link)
            sidebar.appendChild(li)
            if (item.child) {
                var littleUl = document.createElement('ul')
                item.child.forEach(function (littleTitle) {
                    var littleLi = document.createElement('li')
                    var littleLink = document.createElement('a')
                    littleLink.innerHTML = littleTitle.text
                    littleLink.setAttribute('href', '#' + littleTitle.id)
                    littleLink.setAttribute('class', 'markdown-sidebar-link')
                    littleLi.appendChild(littleLink)
                    littleUl.appendChild(littleLi)
                })
                li.appendChild(littleUl)
            }
        }
        settings.element.appendChild(sidebar)
        return map
    }
    if (typeof window !== 'undefined') {
        window.markrunSidebar = markrunSidebar
    }
    if (typeof module !== 'undefined') {
        module.exports = markrunSidebar
    }
})()

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
        sourceLink.innerText = node.innerText + ": " + sourcePath.replace(embedReg, "")
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
document.getElementById("nav").className = "markdown-header"
var markrunSideData = markrunSidebar({
    content: document.getElementById("content"),
    element: document.getElementById("nav")
})
// var topNode = document.createElement('span')
// topNode.innerText = "TOP"
// topNode.className = "gotop"
// document.body.appendChild(topNode)
// var timer = null
// topNode.addEventListener("click", function (){
//     scrollTo(0,0);
// })
