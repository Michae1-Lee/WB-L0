package ui

import (
	"fmt"
	"net/http"
)

type SimpleUi struct{}

func NewSimpleUi() *SimpleUi {
	return &SimpleUi{}
}

func (h *SimpleUi) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>Get Order</title>
</head>
<body>
  <h2>Поиск заказа</h2>
  <form id="orderForm">
    <input type="text" id="uid" placeholder="Введите UID" />
    <button type="submit">Получить</button>
  </form>
  <pre id="result"></pre>

  <script>
    document.getElementById("orderForm").addEventListener("submit", async function(e) {
      e.preventDefault();
      const uid = document.getElementById("uid").value;
      if (!uid) {
        alert("Введите UID");
        return;
      }
      const resp = await fetch("/order/" + uid);
      if (!resp.ok) {
        document.getElementById("result").innerText = "Ошибка: " + resp.status;
        return;
      }
      const data = await resp.json();
      document.getElementById("result").innerText = JSON.stringify(data, null, 2);
    });
  </script>
</body>
</html>
`)
}
