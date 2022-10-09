import express from 'express';
const app = express();

app.post('/', (req, res) => {
  res.send("OK");
});

app.listen(process.env.PORT || 8080, () => {
  console.log('Listening...');
});
