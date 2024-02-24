import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';

function SuiteNoteModal() {

    const {
        selectedSuiteObject,
        print,
        setIsSuiteNoteModalOpen,
        credentials,

    } = useContext(MyContext);

    const [suiteNotes, setSuiteNotes] = useState('');


    const updateSuiteNotes = () => {
        print(suiteNotes);
        print(selectedSuiteObject.suiteUUID);
        print(credentials);
        fetch(`https://www.cs.hmc.edu/~tlam/digitaldraw/api/suites/design`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`
            },
            body: JSON.stringify({
                suiteDesign: suiteNotes,
                suiteUUID: selectedSuiteObject.suiteUUID,
                userJWT: credentials,
            }),
        })
            .then(response => response.json())
            .then(data => {
                if (data.error) {
                    console.log(data.error);
                }
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
                    <button className="delete" aria-label="close" onClick={() => setIsSuiteNoteModalOpen(false)}></button>
                </header>
                <section className="modal-card-body">
                    <textarea
                        className="textarea"
                        placeholder="Enter information about genderlocking, suite culture, etc. here."
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