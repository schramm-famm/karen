const chai = require('chai');
const chaiHttp = require('chai-http');

chai.use(chaiHttp);
const { expect } = chai;

const karenEndpoint = process.env.HOST;
const createdUsers = new Set();

const preExistingUser = {
  name: 'Pre-existing User',
  email: 'preexistinguser@foo.bar',
  password: 'foobar',
};

// Create a pre-existing user that each test can use as it needs.
beforeEach(async () => {
  const res = await chai.request(karenEndpoint)
    .post('/karen/v1/users')
    .send(preExistingUser);

  expect(res).to.have.status(201);
  preExistingUser.id = res.body.id;
  createdUsers.add(res.body.id);
});

// Delete any leftover created users to allow for a fresh starting point for the
// next test.
afterEach(async () => {
  const requests = [];
  Array.from(createdUsers).forEach((userID) => {
    requests.push(chai.request(karenEndpoint)
      .delete('/karen/v1/users/self')
      .set('User-ID', userID));
    createdUsers.delete(userID);
  });

  const responses = await Promise.all(requests);
  responses.forEach((res) => {
    expect(res).to.have.status(204);
  });
});

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
    expect(res.body).to.have.all.keys('id', 'name', 'email');
    expect(res.body).to.have.property('name', 'Foo Bar');
    expect(res.body).to.have.property('email', 'foo@bar.baz');
    createdUsers.add(res.body.id);
  });

  it('should fail when re-using an email', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users')
      .send({
        name: 'new',
        email: preExistingUser.email,
        password: 'password',
      });

    expect(res).to.have.status(409);
  });

  it('should fail when a mandatory field is missing', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users/auth')
      .send({
        name: 'Foo Bar',
        email: 'foo@bar.baz',
      });

    expect(res).to.have.status(400);
  });

  it('should fail when sending a malformed request body', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users')
      .send('this is bad');

    expect(res).to.have.status(400);
  });
});

describe('GET /karen/v1/users/self', () => {
  it('should retrieve the session user', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('id', 'name', 'email');
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
  });

  it('should retrieve the session user with specified fields', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .query({ includes: ['name', 'email'] })
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email');
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
  });

  it('should fail when the user does not exist', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', 999999999);

    expect(res).to.have.status(404);
  });

  it('should fail when sending a bad query', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .query({ includes: ['name', 'email', 'foo'] })
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(400);
  });
});

describe('GET /karen/v1/users/{user_id}', () => {
  it('should retrieve the specified user', async () => {
    const res = await chai.request(karenEndpoint)
      .get(`/karen/v1/users/${preExistingUser.id}`)
      .set('User-ID', 123);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('id', 'name', 'email');
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
  });
});

describe('PATCH /karen/v1/users/self', () => {
  it('should update a user', async () => {
    const res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id)
      .send({
        name: 'New Name',
        email: 'newemail@foo.bar',
        password: 'newpassword',
        avatar_url: 'example.com/newavatar.png',
      });

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', 'New Name');
    expect(res.body).to.have.property('email', 'newemail@foo.bar');
    expect(res.body).to.have.property('avatar_url', 'example.com/newavatar.png');
  });

  it('should update only the specified field(s)', async () => {
    const res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id)
      .send({
        name: 'New Name',
      });

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name');
    expect(res.body).to.have.property('name', 'New Name');
  });

  it('should fail when the user does not exist', async () => {
    const res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', 999999999)
      .send({
        name: 'New Name',
        email: 'newemail@foo.bar',
        password: 'newpassword',
      });

    expect(res).to.have.status(404);
  });

  it('should fail when sending a malformed request body', async () => {
    const res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id)
      .send('this is bad');

    expect(res).to.have.status(400);
  });
});

describe('DELETE /karen/v1/users/self', () => {
  it('should delete the session user', async () => {
    const res = await chai.request(karenEndpoint)
      .delete('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(204);
    createdUsers.delete(preExistingUser.id);
  });

  it('should fail when the user does not exist', async () => {
    const res = await chai.request(karenEndpoint)
      .delete('/karen/v1/users/self')
      .set('User-ID', 999999999);

    expect(res).to.have.status(404);
  });
});

describe('POST /karen/v1/users/auth', () => {
  it('should authenticate correct credentials', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users/auth')
      .send({
        email: preExistingUser.email,
        password: preExistingUser.password,
      });

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('id', 'name', 'email');
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
  });

  it('should catch an incorrect password', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users/auth')
      .send({
        email: preExistingUser.email,
        password: 'wrongpass',
      });

    expect(res).to.have.status(401);
  });

  it('should fail when the user does not exist', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users/auth')
      .send({
        email: 'thisdoesnotexist@sympatico.com',
        password: 'doesn\'t matter',
      });

    expect(res).to.have.status(404);
  });

  it('should fail when a mandatory field is missing', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users/auth')
      .send({
        email: preExistingUser.email,
      });

    expect(res).to.have.status(400);
  });

  it('should fail when sending a malformed request body', async () => {
    const res = await chai.request(karenEndpoint)
      .post('/karen/v1/users/auth')
      .send('this is bad');

    expect(res).to.have.status(400);
  });
});
