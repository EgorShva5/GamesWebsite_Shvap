let b1 = document.getElementById('btn1')
let b2 = document.getElementById('btn2')
let b3 = document.getElementById('btn3')
let b4 = document.getElementById('btn4')
let b5 = document.getElementById('btn5')

const Page = parseInt(b3.value.split('/')[0])
const MaxPage = parseInt(b3.value.split('/')[1])

document.addEventListener('DOMContentLoaded', function() {
    b1.setAttribute("value", "<<")
    b5.setAttribute("value", ">>")

    b2.style.visibility = (Page === 1) ? 'hidden' : 'visible'
    b4.style.visibility = (Page === MaxPage) ? 'hidden' : 'visible'

    b2.setAttribute("value", Page - 1)
    b4.setAttribute("value", Page + 1)
    }
);

b1.onclick = function() {
    window.location.href = window.location.origin + "/catalog?page=1"
}
b2.onclick = function() {
    window.location.href = window.location.origin + "/catalog?page=" + (Page-1)
}
b4.onclick = function() {
    window.location.href = window.location.origin + "/catalog?page=" + (Page+1)
}
b5.onclick = function() {
    window.location.href = window.location.origin + "/catalog?page=" + MaxPage
}