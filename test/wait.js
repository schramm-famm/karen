const request = require('request');

const karenEndpoint = process.env.HOST;

function ping(endpoint, numRetries, timeout, callback) {
  request(endpoint, (err, res) => {
    if (err) {
      return callback(err);
    }

    if (res.statusCode !== 404) {
      console.log('Failed.');
      if (numRetries <= 0) {
        return callback(new Error('Failed too many times.'));
      }

      console.log(`Retrying (${numRetries} more attempts)...`);
      return setTimeout(() => {
        ping(endpoint, numRetries - 1, timeout, callback);
      }, timeout);
    }

    return callback();
  });
}

console.log('Waiting for karen to start up...');
ping(`http://${karenEndpoint}`, 20, 1000, (err) => {
  if (err) {
    console.log(err.message);
    process.exit(1);
  }
  console.log('Success!');
});
