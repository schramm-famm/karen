const chai = require('chai');
const chaiHttp = require('chai-http');

chai.use(chaiHttp);
const { expect } = chai;

const karenEndpoint = process.env.HOST;
const createdUsers = [];

/*
beforeEach(async () => {
  const res = await chai.request(karenEndpoint)
    .post('/karen/v1/users')
    .send({
      name: 'Pre-existing User',
      email: 'preexistinguser@foo.bar',
      password: 'foobar',
    });

  expect(res).to.have.status(201);
  createdUsers.push(res.body.id);
});

afterEach(async () => {
  const requests = [];
  while (createdUsers.length > 0) {
    const userID = createdUsers.pop();
    requests.push(chai.request(karenEndpoint)
      .delete('/karen/v1/users')
      .set('User-ID', userID));
  }

  const responses = await Promise.all(requests);
  responses.forEach((res) => {
    expect(res).to.have.status(204);
  });
});
*/

describe('POST /karen/v1/users', () => {
  it('should create a user', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users')
      .send({
        name: 'Foo Bar',
        email: 'foo@bar.baz',
        password: 'foobarbaz',
      });

    expect(res).to.have.status(201);
    expect(res.body).to.have.property('id');
    expect(res.body).to.have.property('name', 'Foo Bar');
    expect(res.body).to.have.property('email', 'foo@bar.baz');
    createdUsers.push(res.body.id);
  });

  it.skip('should fail when re-using an email', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users')
      .send({
        name: 'new',
        email: 'preexistinguser@foo.bar',
        password: 'password',
      });

    expect(res).to.have.status(409);
  });
});
