document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const errorMessage = document.getElementById('errorMessage');

    try {
        const response = await fetch('http://localhost:3100/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });

        if (!response.ok) {
            throw new Error('Login failed');
        }

        const data = await response.json();
        console.log('Login successful:', data);
        // Handle successful login here (e.g., redirect to dashboard)

    } catch (error) {
        errorMessage.style.display = 'block';
        errorMessage.textContent = 'Login failed. Please check your credentials.';
        console.error('Error:', error);
    }
});
