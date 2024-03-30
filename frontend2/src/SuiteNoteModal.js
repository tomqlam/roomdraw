import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';
import { Browser } from '@syncfusion/ej2-base';
import { ImageEditorComponent } from '@syncfusion/ej2-react-image-editor';
import './App.css';
import Compressor from 'compressorjs';

import * as ReactDOM from 'react-dom';

function SuiteNoteModal() {

    const {
        selectedSuiteObject,
        print,
        setIsSuiteNoteModalOpen,
        credentials,
        setRefreshKey,
        suiteDimensions,
        handleErrorFromTokenExpiry,
  

    } = useContext(MyContext);

    const [suiteNotes, setSuiteNotes] = useState('');
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingClearNotes, setLoadingClearNotes] = useState(false);

    useEffect(() => {
        if (selectedSuiteObject) {
            setSuiteNotes(selectedSuiteObject.suiteDesign);
        }
    }, []);



    const updateSuiteNotes = (notes) => {
        let imageData = imgObj.current.getImageData();
        const canvas = document.createElement('canvas');
        canvas.width = imageData.width;
        canvas.height = imageData.height;
        const context = canvas.getContext('2d');
        context.putImageData(imageData, 0, 0);

        // Convert canvas to blob
        canvas.toBlob((blob) => {
            // Compress the blob
            new Compressor(blob, {
                quality: 0.5,
                success: (compressedResult) => {
                    // Store the compressed blob in state
                    setImage(compressedResult);

                    const url = `https://www.cs.hmc.edu/~tlam/digitaldraw/api/suites/design/${selectedSuiteObject.suiteUUID}`;
                    const formData = new FormData();
                    formData.append('suite_design', compressedResult, 'suite_design.jpg');

                    fetch(url, {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${localStorage.getItem('jwt')}`
                        },
                        body: formData,
                    })
                        .then(response => response.json())
                        .then(data => {
                            if (data.error) {
                                if (handleErrorFromTokenExpiry(data)) {
                                    return;
                                };
                            } else {
                                // updated suite successfully 
                                setIsSuiteNoteModalOpen(false);
                                setRefreshKey(prevKey => prevKey + 1);
                                // commented console.log ("refreshing");
                            }
                        })
                        .catch((error) => {
                            console.error('Error:', error);
                        });
                },
            });
        }, 'image/jpeg');
    }


    const deleteSuiteNotes = (notes) => {
        fetch(`https://www.cs.hmc.edu/~tlam/digitaldraw/api/suites/design/remove/${selectedSuiteObject.suiteUUID}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`
            },
        })
            .then(response => response.json())
            .then(data => {
                if (data.error) {
                    // commented console.log (data.error);
                } else {
                    // updated suite successfully 
                    setIsSuiteNoteModalOpen(false);
                    setRefreshKey(prevKey => prevKey + 1);

                }
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }


    const imgObj = React.useRef(null);

    // const imageEditorCreated = () => {
    //     if (Browser.isDevice) {
    //       imgObj.current.open('https://digitaldraw.b-cdn.net/suite_designs/d589f55c-0a89-42ba-bce8-fe3815a62d0e.jpg');
    //     } else {
    //       imgObj.current.open('https://digitaldraw.b-cdn.net/suite_designs/d589f55c-0a89-42ba-bce8-fe3815a62d0e.jpg');
    //     }
    //   }

    useEffect(() => {
        if (imgObj.current) {
            // const width = suiteDimensions.width * 3; // replace with your desired width
            // const height = suiteDimensions.height * 3; // replace with your desired height

            // // Create a new canvas element
            // const canvas = document.createElement('canvas');

            // // Set the width and height of the canvas
            // canvas.width = width;
            // canvas.height = height;

            // // Get the 2D rendering context for the canvas
            // const ctx = canvas.getContext('2d');

            // // Fill the canvas with white color
            // ctx.fillStyle = 'black';
            // ctx.fillRect(0, 0, width, height);

            // // Set the color for the text
            // ctx.fillStyle = 'white';

            // // Set the font for the text
            // ctx.font = '30px Arial';

            // // Add the text to the canvas
            // ctx.fillText('Some text', 50, 50);

            // // Convert the canvas to a data URL
            // const url = canvas.toDataURL('image/png');
            if (selectedSuiteObject.suiteDesign) {
                const url = selectedSuiteObject.suiteDesign;
                // commented console.log (selectedSuiteObject);

                // Open the image in the ImageEditorComponent
                imgObj.current.open(url);
            }

        }
    }, []);
    const [image, setImage] = React.useState(null);

    const handleSave = () => {
        let imageData = imgObj.current.getImageData();
        const canvas = document.createElement('canvas');
        canvas.width = imageData.width;
        canvas.height = imageData.height;
        const context = canvas.getContext('2d');
        context.putImageData(imageData, 0, 0);
        let base64String = canvas.toDataURL(); // For further usage

        // Convert base64 to raw binary data held in a string
        let byteString = atob(base64String.split(',')[1]);

        // Separate out the mime component
        let mimeString = base64String.split(',')[0].split(':')[1].split(';')[0];

        // Write the bytes of the string to a typed array
        let ia = new Uint8Array(byteString.length);
        for (let i = 0; i < byteString.length; i++) {
            ia[i] = byteString.charCodeAt(i);
        }

        let blob = new Blob([ia], { type: mimeString });

        // Store the blob in state
        setImage(blob);
    };



    return (
        <div className="modal is-active">
            <div className="modal-background"></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <p className="modal-card-title">Update suite notes</p>
                    <button className="delete" aria-label="close" onClick={() => setIsSuiteNoteModalOpen(false)}></button>
                </header>
                <section className="modal-card-body">
                    <p>First you must upload any picture, then crop & overlay text on top!</p>
                    <p>Please do not submit inappropriate pictures, or pictures too thin/tall.</p> <br/>
                    {/* <input type="file" id="fileUpload" /> */}
                    <div id="container" style={{ width: '100%', height: '50vh' }}>
                        <ImageEditorComponent toolbar={['Crop', 'Transform', 'Annotate', 'Image', 'ZoomIn', 'ZoomOut',]} ref={imgObj} />
                    </div>
                    {/* <button onClick={handleSave}>Save</button> */}
                    {/* <textarea
                        className="textarea"
                        placeholder="Enter information about genderlocking, suite culture, etc. here."
                        value={suiteNotes}
                        onChange={event => setSuiteNotes(event.target.value)}
                    /> */}
                </section>
                <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <button className={`button is-primary ${loadingSubmit && "is-loading"}`} onClick={() => {
                        setLoadingSubmit(true);
                        updateSuiteNotes(suiteNotes);
                    }}>Submit</button>
                    <button className={`button is-danger ${loadingClearNotes && "is-loading"}`} onClick={() => {
                        setLoadingClearNotes(true);
                        deleteSuiteNotes();
                    }}>Delete all notes</button>
                </footer>
            </div>
        </div>
    );
}

export default SuiteNoteModal;