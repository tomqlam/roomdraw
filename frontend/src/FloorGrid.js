import 'bulma/css/bulma.min.css';
import React, { createRef, useContext, useEffect, useRef, useState } from 'react';
import { MyContext } from './MyContext';

function FloorGrid({ gridData })
{
    const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
    const divRefs = useRef(gridData.suites.map(() => createRef()));
    const divRef = useRef(null);

    useEffect(() =>
    {
        const handleResize = () =>
        {
            setIsMobile(window.innerWidth <= 768);
        };

        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    const {
        print,
        setIsModalOpen,
        setSelectedItem,
        selectedOccupants,
        setSuiteDimensions,
        isSuiteNoteModalOpen,
        setSelectedOccupants,
        setSelectedSuiteObject,
        getNameById,
        setSelectedRoomObject,
        setPullMethod,
        cellColors,
        selectedID,
        userID,
        onlyShowBumpableRooms,
        userMap,
        dormMapping,
        selectedRoomObject,
        setIsFroshModalOpen,
        setIsSuiteNoteModalOpen,
        selectedPalette,
        roomRefs,
        activeTab,
        isDarkMode
    } = useContext(MyContext);

    function capitalizeFirstLetterOfEachWord(str)
    {
        return str.split(' ')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    async function getOccupantsByRoomNumber(roomNumber)
    {
        return new Promise((resolve, reject) =>
        {
            // given a room number, return the occupants of the room
            // Iterate over each suite
            for (let suite of gridData.suites)
            {
                // Find the room with the given room number within the current suite
                const room = suite.rooms.find(r => r.roomNumber.toString() === roomNumber.toString());

                // If the room exists, resolve the Promise with the list of occupants and the room object
                if (room)
                {
                    setSelectedSuiteObject(suite);
                    print(room.occupant1.toString());
                    print(userMap);
                    resolve({
                        occupants: [
                            room.occupant1 !== 0 ? { value: room.occupant1.toString(), label: `${userMap[room.occupant1].FirstName} ${userMap[room.occupant1].LastName}` } : '',
                            room.occupant2 !== 0 ? { value: room.occupant2.toString(), label: `${userMap[room.occupant2].FirstName} ${userMap[room.occupant2].LastName}` } : '',
                            room.occupant3 !== 0 ? { value: room.occupant3.toString(), label: `${userMap[room.occupant3].FirstName} ${userMap[room.occupant3].LastName}` } : '',
                            room.occupant4 !== 0 ? { value: room.occupant4.toString(), label: `${userMap[room.occupant4].FirstName} ${userMap[room.occupant4].LastName}` } : '',
                        ],
                        roomObject: room
                    });
                    return;
                }
            }
            // commented console.log ("Did not find the occupants");
            // If the room does not exist in any suite, resolve the Promise with an empty array and null
            resolve({
                occupants: ['', '', '', ''],
                roomObject: null
            });
        });
    }

    // each cell in floorgrid
    const gridItemStyle = {
        borderRadius: '4px',
        padding: '8px 12px',
        textAlign: 'center',
        fontSize: '0.95rem',
        color: isDarkMode ? '#f5f5f5' : '#2a2a2a',
        fontWeight: '500',
        overflow: 'hidden',
        whiteSpace: 'nowrap',
        textOverflow: 'ellipsis',
        transition: 'all 0.2s ease',
        cursor: 'pointer',
        userSelect: 'none',
    };

    // entire collection of cells
    const gridContainerStyle = {
        display: 'grid',
        gridTemplateColumns: isMobile ?
            `45px 65px ${getOccupantColumns()}` :
            `70px 150px 1fr 1fr 1fr ${(activeTab === 'Atwood' || activeTab === 'Drinkward' || activeTab === 'Case') ? '1fr' : ''} ${activeTab === 'Case' ? '1fr' : ''}`,
        gap: '4px',
        maxWidth: '900px',
        margin: '0 auto',
        background: 'var(--grid-container-bg)',
        padding: '4px',
        borderRadius: '8px',
    };

    // Helper function to determine occupant columns based on dorm type
    function getOccupantColumns()
    {
        if (activeTab === 'Case')
        {
            return '1fr 1fr 1fr 1fr'; // 4 occupants
        } else if (activeTab === 'Atwood' || activeTab === 'Drinkward')
        {
            return '1fr 1fr 1fr'; // 3 occupants
        } else
        {
            return '1fr 1fr'; // 2 occupants
        }
    }

    // darkens given color by a factor, using match
    function darken(color, factor)
    {
        const f = parseInt(factor, 10) || 0;
        const RGB = color.substring(1).match(/.{2}/g);
        const newColor = RGB.map((c) =>
        {
            const hex = Math.max(0, Math.min(255, parseInt(c, 16) - f)).toString(16);
            return hex.length === 1 ? `0${hex}` : hex;
        });
        return `#${newColor.join('')}`;
    }

    // Lighten a color - useful for dark mode
    function lighten(color, factor)
    {
        const f = parseInt(factor, 10) || 0;
        const RGB = color.substring(1).match(/.{2}/g);
        const newColor = RGB.map((c) =>
        {
            const hex = Math.max(0, Math.min(255, parseInt(c, 16) + f)).toString(16);
            return hex.length === 1 ? `0${hex}` : hex;
        });
        return `#${newColor.join('')}`;
    }

    // Choose appropriate color adjustment based on mode
    function adjustColor(color, factor)
    {
        return isDarkMode ? lighten(color, factor) : darken(color, factor);
    }

    const updateSuiteNotes = (room, ref) =>
    {
        getOccupantsByRoomNumber(room);
        setIsSuiteNoteModalOpen(true);
        // commented console.log (ref.current);
        setSuiteDimensions({
            width: ref.current.offsetWidth,
            height: ref.current.offsetHeight
        });


    }
    // given parameters, return grid item style with correct background color shading
    const getGridItemStyle = (room, occupancy, maxOccupancy, suiteIndex, pullPriority) =>
    {
        // Not valid for pulling
        if (occupancy < maxOccupancy || !userMap || !userMap[selectedID])
        {
            return {
                ...gridItemStyle,
                backgroundColor: selectedPalette.unbumpableRoom,
                color: isDarkMode ? '#ffffff' : '#2a2a2a',
            };
        }

        // Current user lives in this room
        if (userID && userMap && userMap[userID] && room.roomUUID === userMap[userID].RoomUUID)
        {
            return {
                ...gridItemStyle,
                backgroundColor: selectedPalette.currentUserRoom,
                color: isDarkMode ? '#ffffff' : '#2a2a2a',
            };
        }

        // Selected person lives in this room
        if (room.roomUUID === userMap[selectedID].RoomUUID)
        {
            return {
                ...gridItemStyle,
                backgroundColor: selectedPalette.selectedUserRoom,
                color: isDarkMode ? '#ffffff' : '#2a2a2a',
            };
        }

        let backgroundColor = (suiteIndex % 2 === 0 ? selectedPalette.evenSuite : selectedPalette.oddSuite);
        if (pullPriority.isPreplaced)
        {
            backgroundColor = adjustColor(backgroundColor, 50);
        }
        if (!checkBumpable(pullPriority) && onlyShowBumpableRooms)
        {
            backgroundColor = adjustColor(backgroundColor, 50);
        }

        return {
            ...gridItemStyle,
            backgroundColor,
            color: isDarkMode ? '#ffffff' : '#2a2a2a',
        };
    };

    const roomNumberStyle = {
        ...gridItemStyle,
        backgroundColor: selectedPalette.roomNumber,
        fontWeight: '600',
        fontSize: '0.9rem',
        letterSpacing: '0.02em',
        cursor: 'default',
        color: isDarkMode ? '#ffffff' : '#2a2a2a',
    };

    const pullMethodStyle = {
        ...gridItemStyle,
        backgroundColor: selectedPalette.pullMethod,
        cursor: 'default',
        color: isDarkMode ? '#ffffff' : '#2a2a2a',
    };

    const handleCellClick = async (roomNumber) =>
    {
        setSelectedItem(roomNumber);
        // commented console.log ("Room number: " + roomNumber);
        const { occupants, roomObject } = await getOccupantsByRoomNumber(roomNumber);
        setSelectedOccupants(occupants);
        setSelectedRoomObject(roomObject);
        // commented console.log ("has frosh?");
        // commented console.log (roomObject);
        // commented console.log (roomObject.hasFrosh);
        if (roomObject && roomObject.hasFrosh)
        {
            setIsFroshModalOpen(true);
        } else
        {
            // commented console.log (occupants);
            setPullMethod("Pulled themselves");
            setIsModalOpen(true);
        }
    };

    function getPullMethodByRoomNumber(roomNumber)
    {
        // Iterate over each suite
        for (let suite of gridData.suites)
        {
            // Find the room with the given room number within the current suite
            const room = suite.rooms.find(r => r.roomNumber.toString() === roomNumber.toString());

            // If the room exists, return the list of occupants


            if (room)
            {
                if (room.hasFrosh)
                {
                    return "Frosh";
                }

                if (room.pullPriority.pullType === 3)
                {
                    return "Lock Pull";
                }
                var pullPriority = room.pullPriority;
                var finalString = "";
                if (pullPriority.inherited.valid)
                {
                    pullPriority = pullPriority.inherited;
                }
                if (pullPriority.isPreplaced)
                {
                    let shortestOccupant = null;
                    // // commented console.log (room)
                    const roomOccupants = [room.occupant1, room.occupant2, room.occupant3, room.occupant4];

                    roomOccupants.forEach(occupant =>
                    {
                        if (occupant !== 0)
                        {
                            if (userMap[occupant].ReslifeRole !== 'none')
                            {
                                if (shortestOccupant === null || userMap[occupant].ReslifeRole.length < shortestOccupant.length)
                                {
                                    shortestOccupant = userMap[occupant].ReslifeRole;
                                }
                            }
                        }
                    });

                    if (shortestOccupant !== null)
                    {
                        return capitalizeFirstLetterOfEachWord(shortestOccupant);
                    } else
                    {
                        return "Preplaced";
                    }
                }
                if (pullPriority.hasInDorm)
                {
                    finalString += `In-Dorm ${pullPriority.drawNumber}`;

                } else
                {
                    const yearMapping = ["", "", "Sophomore", "Junior", "Senior"];
                    finalString += `${yearMapping[pullPriority.year]} ${pullPriority.drawNumber !== 0 ? pullPriority.drawNumber : ''}`;

                }
                if (room.pullPriority.pullType === 4)
                {
                    return finalString += " (2nd best #)";
                }

                return finalString += `${room.pullPriority.pullType === 2 ? " Pull" : ''}`;
            }
        }

        // If the room does not exist in any suite, return an empty array
        return 'n/a';
    }

    const checkBumpable = (pullPriority) =>
    {
        if (!userMap[selectedID])
        {
            return false; // catching case where the screen hasnt loaded yet
        }
        if (!pullPriority.valid)
        {
            // You can bump this
            return true;
        }
        if (pullPriority.isPreplaced)
        {
            // You can't bump this 
            print("preplaced");

            return false;
        }
        if (pullPriority.pullType === 3)
        {
            return false; // lock pull, cannot bump
        }
        // if inherited, use that pullPriority instead
        if (pullPriority.inherited.valid)
        {
            pullPriority = pullPriority.inherited;
        }
        if (pullPriority.hasInDorm)
        {
            if (!userMap[selectedID].InDorm)
            {
                return false;
            }
        }
        // just compare the numbers
        const yearMapping = {
            "sophomore": 2,
            "junior": 3,
            "senior": 4
        };

        if (yearMapping[userMap[selectedID].Year] < pullPriority.year)
        {
            return false;
        } else if (yearMapping[userMap[selectedID].Year] > pullPriority.year)
        {
            // you are older year
            return true;
        }
        return userMap[selectedID].DrawNumber <= pullPriority.drawNumber;

    }



    return (
        <div style={gridContainerStyle} className="grid-container">
            <div style={roomNumberStyle} className="grid-cell">
                <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Room #</strong>
            </div>
            <div style={roomNumberStyle} className="grid-cell">
                <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Pull</strong>
            </div>
            {!isMobile && (
                <div style={roomNumberStyle} className="grid-cell">
                    <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Suite</strong>
                </div>
            )}
            <div style={roomNumberStyle} className="grid-cell">
                <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Occupant 1</strong>
            </div>
            <div style={roomNumberStyle} className="grid-cell">
                <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Occupant 2</strong>
            </div>
            {((activeTab === 'Atwood' || activeTab === 'Drinkward') || activeTab === 'Case') &&
                <div style={roomNumberStyle} className="grid-cell">
                    <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Occupant 3</strong>
                </div>}
            {activeTab === 'Case' &&
                <div style={roomNumberStyle} className="grid-cell">
                    <strong style={{ color: isDarkMode ? '#ffffff' : 'inherit' }}>Occupant 4</strong>
                </div>}

            {[...gridData.suites].map((suite, suiteIndex) => (
                suite.rooms.map((room, roomIndex) => (
                    <React.Fragment key={roomIndex}>
                        <div
                            style={getGridItemStyle(room, room.maxOccupancy, 1, suiteIndex, room.pullPriority)}
                            onClick={() => handleCellClick(room.roomNumber)}
                            ref={el => { roomRefs.current[room.roomUUID] = el; }}
                            id={room.roomUUID}
                            className="grid-cell"
                        >
                            {room.roomNumber}
                        </div>
                        <div
                            style={getGridItemStyle(room, room.maxOccupancy, 1, suiteIndex, room.pullPriority)}
                            onClick={() => handleCellClick(room.roomNumber)}
                            className="grid-cell"
                        >
                            {getPullMethodByRoomNumber(room.roomNumber)}
                        </div>
                        {!isMobile && roomIndex === 0 && (
                            <div
                                style={{
                                    ...pullMethodStyle,
                                    gridRow: `span ${suite.rooms.length}`,
                                    backgroundColor: suiteIndex % 2 === 0 ? selectedPalette.evenSuite : selectedPalette.oddSuite,
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                }}
                                ref={divRefs.current[suiteIndex]}
                                onClick={() => updateSuiteNotes(room.roomNumber, divRefs.current[suiteIndex])}
                                className="grid-cell suite-cell"
                            >
                                {suite.suiteDesign && (
                                    <img
                                        src={suite.suiteDesign}
                                        alt={suite.suiteDesign}
                                        style={{
                                            maxWidth: '100%',
                                            maxHeight: '50vh',
                                            objectFit: 'contain',
                                            width: 'auto',
                                            height: 'auto',
                                            borderRadius: '4px',
                                        }}
                                    />
                                )}
                            </div>
                        )}
                        <div
                            style={getGridItemStyle(room, room.maxOccupancy, 1, suiteIndex, room.pullPriority)}
                            onClick={() => handleCellClick(room.roomNumber)}
                            className="grid-cell"
                            data-full-name={room.maxOccupancy >= 1 ? (room.hasFrosh ? 'Frosh' : getNameById(room.occupant1)) : ''}
                        >
                            {room.maxOccupancy >= 1 && (room.hasFrosh ? 'Frosh' : getNameById(room.occupant1))}
                        </div>
                        <div
                            style={getGridItemStyle(room, room.maxOccupancy, 2, suiteIndex, room.pullPriority)}
                            onClick={() => handleCellClick(room.roomNumber)}
                            className="grid-cell"
                            data-full-name={room.maxOccupancy >= 2 ? (room.hasFrosh ? 'Frosh' : getNameById(room.occupant2)) : ''}
                        >
                            {room.maxOccupancy >= 2 && (room.hasFrosh ? 'Frosh' : getNameById(room.occupant2))}
                        </div>
                        {((activeTab === 'Atwood' || activeTab === 'Drinkward') || activeTab === 'Case') && (
                            <div
                                style={getGridItemStyle(room, room.maxOccupancy, 3, suiteIndex, room.pullPriority)}
                                onClick={() => handleCellClick(room.roomNumber)}
                                className="grid-cell"
                                data-full-name={room.maxOccupancy >= 3 ? (room.hasFrosh ? 'Frosh' : getNameById(room.occupant3)) : ''}
                            >
                                {room.maxOccupancy >= 3 && (room.hasFrosh ? 'Frosh' : getNameById(room.occupant3))}
                            </div>
                        )}
                        {activeTab === "Case" && (
                            <div
                                style={getGridItemStyle(room, room.maxOccupancy, 4, suiteIndex, room.pullPriority)}
                                onClick={() => handleCellClick(room.roomNumber)}
                                className="grid-cell"
                                data-full-name={room.maxOccupancy >= 4 ? (room.hasFrosh ? 'Frosh' : getNameById(room.occupant4)) : ''}
                            >
                                {room.maxOccupancy >= 4 && (room.hasFrosh ? 'Frosh' : getNameById(room.occupant4))}
                            </div>
                        )}
                    </React.Fragment>
                ))
            ))}
        </div>
    );
}

export default FloorGrid;
