document.getElementById('bannerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        title: e.target.title.value.trim(),
        description: e.target.description.value.trim(),
        url: e.target.url.value.trim()
    };
    try {
        const response = await fetch(window.location.origin + "/api/newbanner", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(formData),
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