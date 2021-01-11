import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const serverURL: string = process.env.REACT_APP_TPHOST || "http://localhost:4000"
const wsURL: string = process.env.REACT_APP_TPWSHOST || "ws://localhost:4000"

const api = (action: string, route: string, content: any = undefined, onload: any = undefined, isImage: boolean = false) => {
  let xhr = new XMLHttpRequest()
  if (onload) {
    xhr.addEventListener('load', onload)
  }
  xhr.addEventListener('error', () => toast("there was an error with the request"))
  xhr.open(action, serverURL + "/api/" + route)
  if (!isImage) {
    xhr.responseType = 'json'
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
  }
  if (content) {
    xhr.send(JSON.stringify(
      content
    ))
  } else {
    xhr.send()
  }
}

export { api, serverURL, wsURL };