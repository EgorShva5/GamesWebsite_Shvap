document.addEventListener('DOMContentLoaded', function() {
    fetch(window.location.origin + '/api/banners')
    .then(response => response.json())
    .then(banners => {
        const container = document.querySelector(".GamesContainer");
        
        banners.forEach(banner => {
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