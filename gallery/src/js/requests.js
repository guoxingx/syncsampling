
import axios from 'axios'

axios.defaults.withCredentials = true;

axios.interceptors.request.use(
  config => {
    config.data = JSON.stringify(config.data);
    config.headers = {
      // 'Content-Type': 'application/x-www-form-urlencoded'
      'Content-Type': 'multipart/form-data'
    };
    return config;
  },
  err => {
    return Promise.reject(err);
  }
);

var HOST = ''

export function get (url) {
  return new Promise(function (resolve, reject) {
    axios.get(url)
    .then(res => { resolve(res) })
    .catch(err => {
      if (err.response.status == 403) {
        reject(err)
      }
    })
  })
}

if (process.env.NODE_ENV === 'development') {
  axios.defaults.baseURL = 'http://localhost:3000'
}

export function getImages() {
  return get(HOST + '/api/images')
}

export function getImage() {
  return get(HOST + '/api/image')
}
