document.getElementById('logoutButton').addEventListener('click', async (e) => {
    e.preventDefault();

    try {
        const response = await fetch(window.location.origin + "/api/logout", {
            method: "POST",
            credentials: 'include'
        });
        if (response.ok) {
            localStorage.removeItem('jwt_token');
            sessionStorage.removeItem('jwt_token');

            window.location.reload()
        }
    } catch (error) {
        alert('Ошибка сети: ' + error.message);
    }

});