# To Use you have to enter that code to console in your browser and scroll down
```js
var data = []
var z = setInterval(() => {
console.log("grabbing")
var y = document.querySelectorAll("[role=\"row\"] [data-testid=\"internal-track-link\"]");
y.forEach(e => {data.push(e.innerText)})
console.log("Grabbed",y)

},1000)
// After stopping that by runnging
clearInterval(z)
// get json of that
JSON.stringify([...Set(data)])
```
and save that to file `message.json`