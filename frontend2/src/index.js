import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { MyContextProvider } from './MyContext';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { registerLicense } from '@syncfusion/ej2-base';

registerLicense('Ngo9BigBOggjHTQxAR8/V1NAaF1cVWhKYVB3WmFZfVpgdV9FZFZVQmYuP1ZhSXxXdkZjWH9fdXVUQGJcWUU=');



const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <GoogleOAuthProvider clientId="799922760808-g6aeg2c32r429srq3pq9mi50eb97jvmk.apps.googleusercontent.com">
  <MyContextProvider>
  <React.StrictMode>
    <App />
  </React.StrictMode>
  </MyContextProvider>
  </GoogleOAuthProvider>

);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
