let b1 = document.getElementById('btn1')
let b2 = document.getElementById('btn2')
let b3 = document.getElementById('btn3')
let b4 = document.getElementById('btn4')
let b5 = document.getElementById('btn5')

const page = parseInt(b3.value.split('/')[0])
const maxpage = parseInt(b3.value.split('/')[1])
const urlparams = new URLSearchParams(window.location.search)

document.addEventListener('DOMContentLoaded', function() {
    b1.setAttribute("value", "<<")
    b5.setAttribute("value", ">>")

    switch (page) {
        case maxpage:
            b5.setAttribute("disabled", true)
            b4.style.visibility = 'hidden'
            b2.setAttribute("value", page - 1)
            break
        case 1:
            b1.setAttribute("disabled", true)
            b2.style.visibility = 'hidden'
            b4.setAttribute("value", page + 1)
            break
        default:
            b2.setAttribute("value", page - 1)
            b4.setAttribute("value", page + 1)
    }
});

b1.onclick = function() {
    window.location.href = window.location.origin + "/home?page=1"
}
b2.onclick = function() {
    window.location.href = window.location.origin + "/home?page=" + (page-1)
}
b4.onclick = function() {
    window.location.href = window.location.origin + "/home?page=" + (page+1)
}
b5.onclick = function() {
    window.location.href = window.location.origin + "/home?page=" + maxpage
}