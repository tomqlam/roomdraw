import React, { useState, useEffect } from 'react';
import { useContext } from 'react';
import { MyContext } from './MyContext';

const ThemeSelectModal = () => {
    const {
        colorPalettes,
        setIsFroshModalOpen,
        selectedPalette,
        setSelectedPalette,
        setIsThemeModalOpen
    } = useContext(MyContext);

    // Store selectedPalette in local storage whenever it changes
    useEffect(() => {
        localStorage.setItem('selectedPalette', JSON.stringify(selectedPalette));
    }, [selectedPalette]);

    const handlePaletteChange = (event) => {
        setSelectedPalette(colorPalettes.find(palette => palette.name === event.target.value));
    };

    const handleColorChange = (colorKey, event) => {
        setSelectedPalette(prevPalette => ({
            ...prevPalette,
            [colorKey]: event.target.value
        }));
    };

    return (
        <div className="modal is-active">
            <div className="modal-background"></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <p className="modal-card-title">Edit DigiDraw Theme</p>
                    <button className="delete" aria-label="close" onClick={() => setIsThemeModalOpen(false)}></button>
                </header>
                <section className="modal-card-body">
                    <div className="field">
                        <label className="label">Set to a preset palette</label>
                        <div className="control">
                            <div className="select">
                                <select onChange={handlePaletteChange}>
                                    {colorPalettes.map((palette, index) => (
                                        <option key={index} value={palette.name}>{palette.name}</option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </div>
                    <div className="field">
                        <label className="label">Customize Header Row</label>
                        <div className="control">
                            <input className="input" type="color" id="color1" name="color1" value={selectedPalette.roomNumber} onChange={event => handleColorChange('roomNumber', event)}/>
                        </div>
                    </div>

                    <div className="field">
                        <label className="label">Customize Odd Suites</label>
                        <div className="control">
                            <input className="input" type="color" id="color2" name="color2" value={selectedPalette.oddSuite} onChange={event => handleColorChange('oddSuite', event)}/>
                        </div>
                    </div>

                    <div className="field">
                        <label className="label">Customize Even Suites</label>
                        <div className="control">
                            <input className="input" type="color" id="color3" name="color3" value={selectedPalette.evenSuite} onChange={event => handleColorChange('evenSuite', event)}/>
                        </div>
                    </div>

                    <div className="field">
                        <label className="label">Customize UnBumpable Rooms</label>
                        <div className="control">
                            <input className="input" type="color" id="color4" name="color4" value={selectedPalette.unbumpableRoom} onChange={event => handleColorChange('unbumpableRoom', event)}/>
                        </div>
                    </div>

                    <div className="field">
                        <label className="label">Customize My Current Room</label>
                        <div className="control">
                            <input className="input" type="color" id="color5" name="color5" value={selectedPalette.myRoom} onChange={event => handleColorChange('myRoom', event)}/>
                        </div>
                    </div>

                    <div className="field">
                        <label className="label">Customize Pull Method</label>
                        <div className="control">
                            <input className="input" type="color" id="color6" name="color6" value={selectedPalette.pullMethod} onChange={event => handleColorChange('pullMethod', event)}/>
                        </div>
                    </div>
                </section>
                <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <button className="button is-primary" onClick={() => {
                        setIsThemeModalOpen(false);
                    }}>Done</button>

                </footer>
            </div>
        </div>
    );
};

export default ThemeSelectModal;