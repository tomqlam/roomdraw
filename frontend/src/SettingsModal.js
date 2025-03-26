import React, { useCallback, useContext, useEffect, useRef, useState } from 'react';
import { HexColorInput, HexColorPicker } from 'react-colorful';
import { MyContext } from './MyContext';
import './SettingsModal.css';

const SettingsModal = () =>
{
    const {
        colorPalettes,
        darkColorPalettes,
        setIsFroshModalOpen,
        selectedPalette,
        setSelectedPalette,
        setIsSettingsModalOpen,
        onlyShowBumpableRooms,
        setOnlyShowBumpableRooms,
        showFloorplans,
        setShowFloorplans,
        isDarkMode,
        toggleDarkMode
    } = useContext(MyContext);

    const [activeColorPicker, setActiveColorPicker] = useState(null);
    const [draftColors, setDraftColors] = useState(selectedPalette);
    const [isCustomPalette, setIsCustomPalette] = useState(selectedPalette.name === "Custom");
    const [savedCustomPalette, setSavedCustomPalette] = useState(() =>
    {
        const savedPalette = localStorage.getItem('customPalette');
        return savedPalette ? JSON.parse(savedPalette) : colorPalettes.find(p => p.name === "Custom");
    });

    // For dark mode
    const [savedDarkCustomPalette, setSavedDarkCustomPalette] = useState(() =>
    {
        const savedPalette = localStorage.getItem('darkCustomPalette');
        return savedPalette ? JSON.parse(savedPalette) : darkColorPalettes.find(p => p.name === "Custom");
    });

    const activePalettes = isDarkMode ? darkColorPalettes : colorPalettes;

    const timeoutRef = useRef(null);
    const popoverRef = useRef(null);

    // Store selectedPalette in local storage whenever it changes
    useEffect(() =>
    {
        localStorage.setItem('selectedPalette', JSON.stringify(selectedPalette));
    }, [selectedPalette]);

    // Store custom palette in local storage whenever it changes
    useEffect(() =>
    {
        if (isCustomPalette)
        {
            if (isDarkMode)
            {
                localStorage.setItem('darkCustomPalette', JSON.stringify(selectedPalette));
                setSavedDarkCustomPalette(selectedPalette);
            } else
            {
                localStorage.setItem('customPalette', JSON.stringify(selectedPalette));
                setSavedCustomPalette(selectedPalette);
            }
        }
    }, [isCustomPalette, selectedPalette, isDarkMode]);

    // Update draft colors when active palettes changes due to dark mode toggle
    useEffect(() =>
    {
        setDraftColors(selectedPalette);
    }, [isDarkMode, selectedPalette]);

    const handlePaletteChange = (event) =>
    {
        const selectedName = event.target.value;
        if (selectedName === "Custom")
        {
            // Restore the saved custom palette
            const customPalette = isDarkMode ? savedDarkCustomPalette : savedCustomPalette;
            setSelectedPalette(customPalette);
            setDraftColors(customPalette);
            setIsCustomPalette(true);
        } else
        {
            const newPalette = activePalettes.find(palette => palette.name === selectedName);
            setSelectedPalette(newPalette);
            setDraftColors(newPalette);
            setIsCustomPalette(false);
        }
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
            // When modifying a predefined palette, create a custom one
            if (!isCustomPalette)
            {
                const customPalette = {
                    ...draftColors,
                    [colorKey]: color,
                    name: "Custom"
                };
                setSelectedPalette(customPalette);

                if (isDarkMode)
                {
                    setSavedDarkCustomPalette(customPalette);
                    localStorage.setItem('darkCustomPalette', JSON.stringify(customPalette));
                } else
                {
                    setSavedCustomPalette(customPalette);
                    localStorage.setItem('customPalette', JSON.stringify(customPalette));
                }

                setIsCustomPalette(true);
            } else
            {
                // Just update the existing custom palette
                const updatedPalette = {
                    ...selectedPalette,
                    [colorKey]: color
                };
                setSelectedPalette(updatedPalette);

                if (isDarkMode)
                {
                    setSavedDarkCustomPalette(updatedPalette);
                    localStorage.setItem('darkCustomPalette', JSON.stringify(updatedPalette));
                } else
                {
                    setSavedCustomPalette(updatedPalette);
                    localStorage.setItem('customPalette', JSON.stringify(updatedPalette));
                }
            }
        }, 100);
    }, [setSelectedPalette, draftColors, isCustomPalette, selectedPalette, isDarkMode]);

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

    // Handle dark mode toggle
    const handleDarkModeToggle = () =>
    {
        toggleDarkMode();
        // The palette will be automatically switched in the MyContext useEffect
    };

    const ColorPickerPopover = ({ colorKey, color }) =>
    {
        const [position, setPosition] = useState({ top: 0, left: 0 });
        const buttonRef = useRef(null);

        const updatePosition = useCallback(() =>
        {
            if (buttonRef.current)
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
        }, []);

        // Update position when color picker is opened
        useEffect(() =>
        {
            if (activeColorPicker === colorKey)
            {
                updatePosition();
                // Add window resize listener
                window.addEventListener('resize', updatePosition);
                return () => window.removeEventListener('resize', updatePosition);
            }
        }, [activeColorPicker, colorKey, updatePosition]);

        return (
            <div className="color-picker-container">
                <button
                    ref={buttonRef}
                    className="color-swatch"
                    style={{ backgroundColor: color }}
                    onClick={() =>
                    {
                        if (activeColorPicker === colorKey)
                        {
                            setActiveColorPicker(null);
                        } else
                        {
                            setActiveColorPicker(colorKey);
                            // Update position immediately when opening
                            setTimeout(updatePosition, 0);
                        }
                    }}
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
                            left: `${position.left}px`,
                            opacity: position.top === 0 ? 0 : 1,
                            transition: 'opacity 0.2s ease'
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
                <header className="modal-card-head">
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
                        <div className="field mt-4">
                            <label className="checkbox">
                                <input
                                    type="checkbox"
                                    checked={isDarkMode}
                                    onChange={handleDarkModeToggle}
                                    className="mr-2"
                                />
                                <span className="icon-text">
                                    <span className="icon">
                                        <i className={`fas ${isDarkMode ? 'fa-sun' : 'fa-moon'}`}></i>
                                    </span>
                                    <span>{isDarkMode ? 'Light' : 'Dark'} mode</span>
                                </span>
                                <p className="help">Toggle to {isDarkMode ? 'Light' : 'Dark'} mode for the application</p>
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
                                        value={isCustomPalette ? "Custom" : selectedPalette.name}
                                        onChange={handlePaletteChange}
                                    >
                                        {activePalettes.map(palette => (
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
                                    <label className="label is-small">Pull Method</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="pullMethod" color={draftColors.pullMethod} />
                                    </div>
                                </div>
                            </div>

                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">My Current Room</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="currentUserRoom" color={draftColors.currentUserRoom} />
                                    </div>
                                </div>
                            </div>

                            <div className="column is-half">
                                <div className="field">
                                    <label className="label is-small">Selected User Room</label>
                                    <div className="control">
                                        <ColorPickerPopover colorKey="selectedUserRoom" color={draftColors.selectedUserRoom} />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                <footer className="modal-card-foot" style={{ justifyContent: 'flex-end' }}>
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