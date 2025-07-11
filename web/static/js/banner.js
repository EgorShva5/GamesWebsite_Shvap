document.getElementById('bannerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = new FormData(e.target)

    try {
        const response = await fetch(window.location.origin + "/api/newbanner", {
            method: "POST",
            body: formData,
        });
        const result = await response.json();
        
        if (response.ok) {
            alert('Добавлено!!!')
        } else {
            alert(`Ошибка: ${result.error}`);
        }
    } catch (error) {
        alert('Ошибка сети: ' + error.message);
    }

});