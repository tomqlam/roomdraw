import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';
import { ImageEditorComponent } from '@syncfusion/ej2-react-image-editor';
import './App.css';
import * as ReactDOM from 'react-dom';

function SuiteNoteModal() {

    const {
        selectedSuiteObject,
        print,
        setIsSuiteNoteModalOpen,
        credentials,
        setRefreshKey,

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
        const url = `/suites/design/${selectedSuiteObject.suiteUUID}`;

        const formData = new FormData();
        const fileField = document.querySelector('input[type="file"]');

        formData.append('suite_design', fileField.files[0]);

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
                    console.log(data.error);
                } else {
                    // updated suite successfully 
                    setIsSuiteNoteModalOpen(false);
                    setRefreshKey(prevKey => prevKey + 1);
                    console.log("refreshing");
                }
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }


    // const updateSuiteNotes = (notes) => {
    //     fetch(`/suites/design/${selectedSuiteObject.suiteUUID}`, {
    //         method: 'POST',
    //         headers: {
    //             'Content-Type': 'application/json',
    //             'Authorization': `Bearer ${localStorage.getItem('jwt')}`
    //         },
    //         body: JSON.stringify({
    //             SuiteDesign: notes,
    //         }),
    //     })
    //         .then(response => response.json())
    //         .then(data => {
    //             if (data.error) {
    //                 console.log(data.error);
    //             } else {
    //                 // updated suite successfully 
    //                 setIsSuiteNoteModalOpen(false);
    //                 setRefreshKey(prevKey => prevKey + 1);
    //                 console.log("refreshing");

    //             }
    //         })
    //         .catch((error) => {
    //             console.error('Error:', error);
    //         });
    // }


    return (
        <div className="modal is-active">
            <div className="modal-background"></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <p className="modal-card-title">Update suite notes</p>
                    <button className="delete" aria-label="close" onClick={() => setIsSuiteNoteModalOpen(false)}></button>
                </header>
                <section className="modal-card-body">
                    <input type="file" id="fileUpload" />
                    <div id="container" style={{ width: '100%', height: '100%' }}>
    <ImageEditorComponent />
</div>
                    <textarea
                        className="textarea"
                        placeholder="Enter information about genderlocking, suite culture, etc. here."
                        value={suiteNotes}
                        onChange={event => setSuiteNotes(event.target.value)}
                    />
                </section>
                <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <button className={`button is-primary ${loadingSubmit && "is-loading"}`} onClick={() => {
                        setLoadingSubmit(true);
                        updateSuiteNotes(suiteNotes);
                    }}>Submit</button>
                    <button className={`button is-danger ${loadingClearNotes && "is-loading"}`} onClick={() => {
                        setLoadingClearNotes(true);
                        setSuiteNotes('');
                        updateSuiteNotes('');
                    }}>Delete all notes</button>
                </footer>
            </div>
        </div>
    );
}

export default SuiteNoteModal;