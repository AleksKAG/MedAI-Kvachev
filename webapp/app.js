Telegram.WebApp.ready();
Telegram.WebApp.expand();

let selectedFile = null;

document.getElementById('fileInput').addEventListener('change', (e) => {
  selectedFile = e.target.files[0];
});

document.getElementById('uploadBtn').addEventListener('click', async () => {
  const errorDiv = document.getElementById('error');
  const resultsDiv = document.getElementById('results');
  const resultsList = document.getElementById('resultsList');

  errorDiv.classList.add('hidden');
  resultsDiv.classList.add('hidden');

  if (!selectedFile) {
    errorDiv.textContent = 'Выберите файл';
    errorDiv.classList.remove('hidden');
    return;
  }

  const formData = new FormData();
  formData.append('file', selectedFile);

  try {
    const response = await fetch('http://localhost:8081/api/ocr-parse', {
      method: 'POST',
      body: formData,
    });

    const data = await response.json();

    if (!data.success) {
      throw new Error(data.error || 'Ошибка обработки');
    }

    // Показать результаты
    resultsList.innerHTML = '';
    data.results.forEach(r => {
      const div = document.createElement('div');
      div.className = 'result-item';
      div.innerHTML = `<strong>${r.name}</strong>: ${r.value} ${r.unit}<br>
                       <small>${r.interpretation}</small>`;
      resultsList.appendChild(div);
    });

    resultsDiv.classList.remove('hidden');

    // Сохранить данные для отправки
    window.parsedResults = data.results;

  } catch (err) {
    errorDiv.textContent = 'Ошибка: ' + err.message;
    errorDiv.classList.remove('hidden');
  }
});

document.getElementById('saveBtn').addEventListener('click', () => {
  if (!window.parsedResults) return;

  Telegram.WebApp.sendData(JSON.stringify({
    action: 'save_lab_results',
    results: window.parsedResults
  }));

  Telegram.WebApp.close();
});
