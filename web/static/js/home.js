const PerPage = 3
var PageNum = 1

let b1 = document.getElementById('btn1')
let b2 = document.getElementById('btn2')
let b3 = document.getElementById('btn3')
let b4 = document.getElementById('btn4')
let b5 = document.getElementById('btn5')

let Ba 

document.addEventListener('DOMContentLoaded', function() {
    fetch(window.location.origin + '/api/banners')
    .then(response => response.json())
    .then(banners => {
        const container = document.querySelector(".GamesContainer");

        Ba = banners

        let MaxPage = parseInt(banners.length - PerPage)
        let VisibleBan = banners.slice((PageNum - 1) * PerPage, PerPage * PageNum)

        b1.value = '1'
        b2.value = PageNum-1
        b3.value = PageNum
        b4.value = PageNum+1
        b5.value = MaxPage

        VisibleBan.forEach(banner => {
        const bannerHTML = `
            <div class="Game">
                <a href="${banner.url}" title="Привет!">
                <img src="/static/img/ewwie.jpg" alt="Картинка">
                <h1>${banner.title}</h1>
                <p>${banner.description}</p>
                <p>Автор: ${banner.author}</p>
                </a>
            </div>
        `;
        container.insertAdjacentHTML('beforeend', bannerHTML);
        });
    })
    .catch(error => console.error('Ошибка:', error));

})

alert(Ba)