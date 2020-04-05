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
  preExistingUser.avatar_url = '';
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
    const newUser = {
      name: 'Foo Bar',
      email: 'foo@bar.baz',
      password: 'foobarbaz',
    };

    let res = await chai.request(karenEndpoint)
      .post('/karen/v1/users')
      .send(newUser);
    newUser.avatar_url = '';

    expect(res).to.have.status(201);
    createdUsers.add(res.body.id);
    expect(res.body).to.have.all.keys('id', 'name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', newUser.name);
    expect(res.body).to.have.property('email', newUser.email);
    expect(res.body).to.have.property('avatar_url', newUser.avatar_url);

    // Check that the new user resource can be retrieved.
    res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', res.body.id);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', newUser.name);
    expect(res.body).to.have.property('email', newUser.email);
    expect(res.body).to.have.property('avatar_url', newUser.avatar_url);
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
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
    expect(res.body).to.have.property('avatar_url', preExistingUser.avatar_url);
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
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
    expect(res.body).to.have.property('avatar_url', preExistingUser.avatar_url);
  });
});

describe('GET /karen/v1/users', () => {
  it('should retrieve the user specified by email', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users')
      .query({ email: preExistingUser.email })
      .set('User-ID', 123);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('id', 'name', 'email', 'avatar_url');
    expect(res.body).to.have.property('id', preExistingUser.id);
    expect(res.body).to.have.property('name', preExistingUser.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
    expect(res.body).to.have.property('avatar_url', preExistingUser.avatar_url);
  });

  it('should fail when no user with the specified email exists', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users')
      .query({ email: 'nobodyusingthis@hotmail.com' })
      .set('User-ID', 123);

    expect(res).to.have.status(404);
  });

  it('should fail when missing the "email" query parameter', async () => {
    const res = await chai.request(karenEndpoint)
      .get('/karen/v1/users')
      .set('User-ID', 123);

    expect(res).to.have.status(400);
  });
});

describe('PATCH /karen/v1/users/self', () => {
  it('should update a user', async () => {
    const updatedFields = {
      name: 'New Name',
      email: 'newemail@foo.bar',
      password: 'newpassword',
      avatar_url: 'example.com/newavatar.png',
    };

    let res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id)
      .send(updatedFields);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', updatedFields.name);
    expect(res.body).to.have.property('email', updatedFields.email);
    expect(res.body).to.have.property('avatar_url', updatedFields.avatar_url);

    // Check that the user resource gets retrieved with the updated properties.
    res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', updatedFields.name);
    expect(res.body).to.have.property('email', updatedFields.email);
    expect(res.body).to.have.property('avatar_url', updatedFields.avatar_url);
  });

  it('should update only the specified field(s)', async () => {
    const updatedFields = {
      name: 'New Name',
    };

    let res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id)
      .send(updatedFields);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', updatedFields.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
    expect(res.body).to.have.property('avatar_url', preExistingUser.avatar_url);

    // Check that the user resource gets retrieved with the updated property.
    res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', updatedFields.name);
    expect(res.body).to.have.property('email', preExistingUser.email);
    expect(res.body).to.have.property('avatar_url', preExistingUser.avatar_url);
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

  it('should fail when using an email that is already taken', async () => {
    // Create a user to be updated.
    const testUser = {
      name: 'Test Name',
      email: 'testemail@foo.bar',
      password: 'testpassword',
    };

    let res = await chai.request(karenEndpoint)
      .post('/karen/v1/users')
      .send(testUser);

    expect(res).to.have.status(201);
    createdUsers.add(res.body.id);
    testUser.id = res.body.id;
    testUser.avatar_url = '';

    // Update the test user
    const updatedFields = {
      name: 'New Name',
      email: preExistingUser.email,
      password: 'newpassword',
      avatar_url: 'example.com/newavatar.png',
    };

    res = await chai.request(karenEndpoint)
      .patch('/karen/v1/users/self')
      .set('User-ID', testUser.id)
      .send(updatedFields);

    expect(res).to.have.status(409);

    // Check that the user's fields haven't changed.
    res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', testUser.id);

    expect(res).to.have.status(200);
    expect(res.body).to.have.all.keys('name', 'email', 'avatar_url');
    expect(res.body).to.have.property('name', testUser.name);
    expect(res.body).to.have.property('email', testUser.email);
    expect(res.body).to.have.property('avatar_url', testUser.avatar_url);
  });
});

describe('DELETE /karen/v1/users/self', () => {
  it('should delete the session user', async () => {
    let res = await chai.request(karenEndpoint)
      .delete('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(204);
    createdUsers.delete(preExistingUser.id);

    // Check that the user resource is no longer available.
    res = await chai.request(karenEndpoint)
      .get('/karen/v1/users/self')
      .set('User-ID', preExistingUser.id);

    expect(res).to.have.status(404);
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
    expect(res.body).to.have.property('id', preExistingUser.id);
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
