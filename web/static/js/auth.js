let LoginForm = document.getElementById('registerForm')
let EnterForm = document.getElementById('loginForm')
let ChBtn = document.getElementById('ch_btn')
let IsReg = false

ChBtn.addEventListener('click', function () {
    if (!IsReg) {
        LoginForm.style.display = 'none';
        EnterForm.style.display = 'block';
        ChBtn.value = 'Ещё нет аккаунта? Регистрация'; 
    } else {
        LoginForm.style.display = 'block';
        EnterForm.style.display = 'none'; 
        ChBtn.value = 'Уже есть аккаунт? Вход';
    }
    IsReg = !IsReg
})

document.getElementById('registerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        display: e.target.display.value,
        login: e.target.login.value,
        password: e.target.password.value
    };
    try {
        const response = await fetch(window.location.origin + "/api/register", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(formData),
        });
        const result = await response.json();
        
        if (response.ok) {
            alert('Регистрация успешна!');
        } else {
            alert(`Ошибка: ${result.error}`);
        }
    } catch (error) {
        alert('Ошибка сети: ' + error.message);
    }
});
document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        login: e.target.login.value,
        password: e.target.password.value
    };
    try {
        const response = await fetch(window.location.origin + "/api/login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(formData),
        });
        const result = await response.json();
        
        if (response.ok) {
            window.location.href = '/home'
        } else {
            alert(`Ошибка: ${result.error}`);
        }
    } catch (error) {
        alert('Ошибка сети: ' + error.message);
    }
});