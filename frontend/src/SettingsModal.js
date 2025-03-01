import React, { useCallback, useContext, useEffect, useRef, useState } from 'react';
import { HexColorInput, HexColorPicker } from 'react-colorful';
import { MyContext } from './MyContext';
import './SettingsModal.css';

const SettingsModal = () =>
{
    const {
        colorPalettes,
        setIsFroshModalOpen,
        selectedPalette,
        setSelectedPalette,
        setIsSettingsModalOpen,
        onlyShowBumpableRooms,
        setOnlyShowBumpableRooms,
        showFloorplans,
        setShowFloorplans,
    } = useContext(MyContext);

    const [activeColorPicker, setActiveColorPicker] = useState(null);
    const [draftColors, setDraftColors] = useState(selectedPalette);
    const timeoutRef = useRef(null);
    const popoverRef = useRef(null);

    // Store selectedPalette in local storage whenever it changes
    useEffect(() =>
    {
        localStorage.setItem('selectedPalette', JSON.stringify(selectedPalette));
    }, [selectedPalette]);

    const handlePaletteChange = (event) =>
    {
        const newPalette = colorPalettes.find(palette => palette.name === event.target.value);
        setSelectedPalette(newPalette);
        setDraftColors(newPalette);
    };

    const handleColorChange = useCallback((colorKey, color) =>
    {
        setDraftColors(prev => ({
            ...prev,
            [colorKey]: color
        }));

        // Clear any existing timeout
        if (timeoutRef.current)
        {
            clearTimeout(timeoutRef.current);
        }

        // Set a new timeout to update the actual palette
        timeoutRef.current = setTimeout(() =>
        {
            setSelectedPalette(prev => ({
                ...prev,
                [colorKey]: color
            }));
        }, 100);
    }, [setSelectedPalette]);

    // Handle clicking outside of color picker
    useEffect(() =>
    {
        const handleClickOutside = (event) =>
        {
            if (popoverRef.current && !popoverRef.current.contains(event.target))
            {
                // Check if the click was on a color swatch button
                const isColorSwatchClick = event.target.closest('.color-swatch');
                if (!isColorSwatchClick)
                {
                    setActiveColorPicker(null);
                }
            }
        };

        // Only add the listener when a color picker is active
        if (activeColorPicker)
        {
            document.addEventListener('mousedown', handleClickOutside);
            return () => document.removeEventListener('mousedown', handleClickOutside);
        }
    }, [activeColorPicker]);

    // Cleanup timeout on unmount
    useEffect(() =>
    {
        return () =>
        {
            if (timeoutRef.current)
            {
                clearTimeout(timeoutRef.current);
            }
        };
    }, []);

    const ColorPickerPopover = ({ colorKey, color }) =>
    {
        const [position, setPosition] = useState({ top: 0, left: 0 });
        const buttonRef = useRef(null);

        useEffect(() =>
        {
            if (activeColorPicker === colorKey && buttonRef.current)
            {
                const rect = buttonRef.current.getBoundingClientRect();
                const spaceBelow = window.innerHeight - rect.bottom;
                const spaceAbove = rect.top;
                const pickerHeight = 250; // Approximate height of the color picker

                // Position horizontally
                let left = rect.left;
                if (left + 240 > window.innerWidth)
                {
                    left = window.innerWidth - 250;
                }

                // Position vertically - prefer below, but go above if not enough space
                let top;
                if (spaceBelow >= pickerHeight || spaceBelow >= spaceAbove)
                {
                    top = rect.bottom + 5;
                } else
                {
                    top = rect.top - pickerHeight - 5;
                }

                setPosition({ top, left });
            }
        }, [activeColorPicker, colorKey]);

        return (
            <div className="color-picker-container">
                <button
                    ref={buttonRef}
                    className="color-swatch"
                    style={{ backgroundColor: color }}
                    onClick={() => setActiveColorPicker(activeColorPicker === colorKey ? null : colorKey)}
                >
                    <span className="color-value">{color}</span>
                </button>
                {activeColorPicker === colorKey && (
                    <div
                        className="color-picker-popover"
                        ref={popoverRef}
                        style={{
                            position: 'fixed',
                            top: `${position.top}px`,
                            left: `${position.left}px`
                        }}
                    >
                        <HexColorPicker
                            color={color}
                            onChange={(newColor) => handleColorChange(colorKey, newColor)}
                        />
                        <div className="color-input-container">
                            <HexColorInput
                                color={color}
                                onChange={(newColor) => handleColorChange(colorKey, newColor)}
                                prefixed
                            />
                        </div>
                    </div>
                )}
            </div>
        );
    };

    return (
        <div className="modal is-active">
            <div className="modal-background" onClick={() => setIsSettingsModalOpen(false)}></div>
            <div className="modal-card" style={{ maxWidth: '500px' }}>
                <header className="modal-card-head" style={{ background: '#f8f9fa' }}>
                    <p className="modal-card-title">
                        <span className="icon-text">
                            <span className="icon">
                                <i className="fas fa-palette"></i>
                            </span>
                            <span>Visual Settings</span>
                        </span>
                    </p>
                    <button
                        className="delete"
                        aria-label="close"
                        onClick={() => setIsSettingsModalOpen(false)}
                    ></button>
                </header>

                <section className="modal-card-body">
                    <div className="box">
                        <h3 className="title is-5 mb-3">Display Options</h3>
                        <div className="field">
                            <label className="checkbox">
                                <input
                                    type="checkbox"
                                    checked={onlyShowBumpableRooms}
                                    onChange={() => setOnlyShowBumpableRooms(!onlyShowBumpableRooms)}
                                    className="mr-2"
                                />
                                Darken rooms selected person can't pull
                                <p className="help">This will darken preplaced rooms extra</p>
                            </label>
                        </div>
                        <div className="field mt-4">
                            <label className="checkbox">
                                <input
                                    type="checkbox"
                                    checked={showFloorplans}
                                    onChange={() => setShowFloorplans(!showFloorplans)}
                                    className="mr-2"
                                />
                                Show floorplans
                                <p className="help">Display floorplan images next to the room grid</p>
                            </label>
                        </div>
                    </div>

                    <div className="box">
                        <h3 className="title is-5 mb-3">Color Theme</h3>
                        <div className="field">
                            <label className="label">Select a color palette</label>
                            <div className="control">
                                <div className="select is-fullwidth">
                                    <select
                                        value={selectedPalette.name}
                                        onChange={handlePaletteChange}
                                    >
                                        {colorPalettes.map(palette => (
                                            <option key={palette.name} value={palette.name}>
                                                {palette.name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>
                        </div>

                        <div className="columns is-multiline mt-4">
                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">Header Row</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="roomNumber" color={draftColors.roomNumber} />
                                    </div>
                                </div>
                            </div>

                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">Odd Suites</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="oddSuite" color={draftColors.oddSuite} />
                                    </div>
                                </div>
                            </div>

                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">Even Suites</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="evenSuite" color={draftColors.evenSuite} />
                                    </div>
                                </div>
                            </div>

                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">My Current Room</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="selectedUserRoom" color={draftColors.selectedUserRoom} />
                                    </div>
                                </div>
                            </div>

                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">Pull Method</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="pullMethod" color={draftColors.pullMethod} />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                <footer className="modal-card-foot" style={{ background: '#f8f9fa', justifyContent: 'flex-end' }}>
                    <button
                        className="button is-primary"
                        onClick={() => setIsSettingsModalOpen(false)}
                    >
                        <span className="icon">
                            <i className="fas fa-check"></i>
                        </span>
                        <span>Save changes</span>
                    </button>
                </footer>
            </div>
        </div>
    );
};

export default SettingsModal;