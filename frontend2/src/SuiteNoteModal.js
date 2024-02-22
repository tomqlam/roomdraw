import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';

function SuiteNoteModal() {

    const {
        selectedSuiteObject,
    } = useContext(MyContext);

    const [suiteNotes, setSuiteNotes] = useState('');


    const updateSuiteNotes = () => {
        fetch('/suites/notes', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                SuiteDesign: suiteNotes,
                SuiteUUID: selectedSuiteObject.uuid,
            }),
        })
            .then(response => response.json())
            .then(data => {
                console.log('Success:', data);
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }


    return (
        <div className="modal is-active">
            <div className="modal-background"></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <p className="modal-card-title">Update suite notes</p>
                    <button className="delete" aria-label="close" ></button>
                </header>
                <section className="modal-card-body">
                    <textarea
                        className="textarea"
                        placeholder="Enter notes here"
                        value={suiteNotes}
                        onChange={event => setSuiteNotes(event.target.value)}
                    />
                </section>
                <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <button className="button is-primary" onClick={updateSuiteNotes}>Submit</button>
                    <button className="button is-danger" >Delete all notes</button>
                </footer>
            </div>
        </div>
    );
}

export default SuiteNoteModal;