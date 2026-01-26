import { GoogleOAuthProvider } from '@react-oauth/google';
import { registerLicense } from '@syncfusion/ej2-base';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';
import { MyContextProvider } from './context/MyContext';
import reportWebVitals from './reportWebVitals';


registerLicense('Ngo9BigBOggjHTQxAR8/V1NAaF5cWWJCf1FpRmJGdld5fUVHYVZUTXxaS00DNHVRdkdnWXxed3VQRWZeVkZ3XEo=');

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
// to log results (for example: reportWebVitals(// commented console.log ))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
