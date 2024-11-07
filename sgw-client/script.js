document.getElementById('analysisForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const linkedinUrl = document.getElementById('linkedinUrl').value;
    const errorMessage = document.getElementById('errorMessage');
    const loadingSpinner = document.getElementById('loadingSpinner');
    const resultsContainer = document.getElementById('results');

    // Reset display
    errorMessage.style.display = 'none';
    loadingSpinner.style.display = 'block';
    resultsContainer.style.display = 'none';

    try {
        const response = await fetch('http://localhost:3100/api/home', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password, linkedinUrl })
        });

        if (!response.ok) {
            throw new Error('Analysis failed');
        }

        const data = await response.json();
        displayResults(data);

    } catch (error) {
        errorMessage.style.display = 'block';
        errorMessage.textContent = 'Analysis failed. Please try again.';
        console.error('Error:', error);
    } finally {
        loadingSpinner.style.display = 'none';
    }
});

function displayResults(data) {
    const resultsContainer = document.getElementById('results');
    const messageSection = document.getElementById('messageSection');
    const paramsSection = document.getElementById('paramsSection');
    const postsSection = document.getElementById('postsSection');

    // Clear previous results
    messageSection.innerHTML = '';
    paramsSection.innerHTML = '';
    postsSection.innerHTML = '';

    // Display message
    messageSection.innerHTML = `
        <div class="section-title">Message:</div>
        <div>${data.msg}</div>
    `;

    // Display parameters used
    paramsSection.innerHTML = `
        <div class="section-title">Parameters Used:</div>
        <ul class="params-list">
            ${data.paramsUsed.map(param => `<li>${param}</li>`).join('')}
        </ul>
    `;

    // Display posts or no posts message
    postsSection.innerHTML = '<div class="section-title">Posts:</div>';
    if (data.recentPosts) {
        data.recentPosts = JSON.parse(data.recentPosts)
    }
    if (data.recentPosts.length > 0) {
        data.recentPosts.forEach(post => {
            postsSection.innerHTML += `
                <div class="post-item">
                    <div class="post-content">${post.content}</div>
                </div>
            `;
        });
    } else {
        postsSection.innerHTML += '<div>User has no posts</div>';
    }

    // Show results container
    resultsContainer.style.display = 'block';
}
