import React, { createContext, useEffect, useRef, useState } from 'react';


export const MyContext = createContext();

export const MyContextProvider = ({ children }) =>
{
    const [currPage, setCurrPage] = useState('Home'); //TODO DELETE
    const [isModalOpen, setIsModalOpen] = useState(false); // If bump modal is open
    const [rooms, setRooms] = useState([]); // json for all rooms
    const [selectedItem, setSelectedItem] = useState(null); // selected room number (integer)
    const [selectedOccupants, setSelectedOccupants] = useState(['0', '0', '0', '0']); // array of occupants, '0' string for none occupant, number as string for occupant
    const [pullMethod, setPullMethod] = useState('Pulled themselves'); // pull method currently selected in dropdown
    const [showModalError, setShowModalError] = useState(false); // if there is an error upon submitting 
    const [onlyShowBumpableRooms, setOnlyShowBumpableRooms] = useState(false); // toggle darkening nonbumpable rooms
    const [gridData, setGridData] = useState([]); // all coalesced data for every dorm
    const [userMap, setUserMap] = useState(null); // information about all users 
    const [selectedRoomObject, setSelectedRoomObject] = useState(null); // json object for current room
    const [selectedSuiteObject, setSelectedSuiteObject] = useState(null); // json object for current suite
    const [refreshKey, setRefreshKey] = useState(0); // key, when incremented, refreshes the main page
    const [pullError, setPullError] = useState("There was an unknown error. Please try again."); // text of error showig up when you can't pull
    const [credentials, setCredentials] = useState(null); // jwt token for user
    const [lastRefreshedTime, setLastRefreshedTime] = useState(new Date()); // last time the page was refreshed
    const [isSuiteNoteModalOpen, setIsSuiteNoteModalOpen] = useState(false); // If suite note modal 
    const [isFroshModalOpen, setIsFroshModalOpen] = useState(false); // If frosh modal is open
    const [suiteDimensions, setSuiteDimensions] = useState({ width: 0, height: 0 }); // dimensions of the suite
    const [isSettingsModalOpen, setIsSettingsModalOpen] = useState(false); // If theme modal is open
    const [isUserSettingsModalOpen, setIsUserSettingsModalOpen] = useState(false); // If user settings modal is open
    const roomRefs = useRef({}); // references to all the room divs
    const [showFloorplans, setShowFloorplans] = useState(false);
    const [myRoom, setMyRoom] = useState("Unselected"); // to show what room current logged in user is in

    const roomNumbers = ["101", "Q1D", "118", "201", "Q2C", "Q2D", "218"]; // Fill this array with the suite UUIDs you want to split
    const floorNames = ["LRL (Topless)", "LLL", "URL", "ULL"]; // Fill this array with the custom floor names

    // Initialize active tab state from localStorage or default to 'Atwood'

    const adminList = ["smao@g.hmc.edu", "tlam@g.hmc.edu"]

    const getRoomUUIDFromUserID = (userID) =>
    {
        if (rooms)
        {
            for (let room of rooms)
            {

                if (room.Occupants && room.Occupants.includes(Number(userID)))
                {
                    // they are this room

                    return room.RoomUUID;
                }
            }


        }
        return null;
    }

    const [activeTab, setActiveTab] = useState(() =>
    {
        const savedTab = localStorage.getItem('activeTab');
        return savedTab !== null ? savedTab : 'Atwood';
    });

    const [selectedID, setSelectedID] = useState(() =>
    {
        const selectedID = localStorage.getItem('selectedID');
        return selectedID !== null ? selectedID : '8'; //TODO 
    });

    const [userID, setUserID] = useState(() =>
    {
        const userID = localStorage.getItem('userID');
        return userID !== null ? userID : '-1'; //TODO 
    });

    const handleErrorFromTokenExpiry = (data) =>
    {
        if (data.error === "Invalid token")
        {
            setCredentials(null);
            localStorage.removeItem('jwt');
            return true;
        }
        return false;
    }

    useEffect(() =>
    {
        const interval = setInterval(() =>
        {
            if (credentials && !document.hidden)
            {
                setRefreshKey(prevKey => prevKey + 1);
                setLastRefreshedTime(new Date());
                // commented console.log ("refreshed ONE DORM");
            }
        }, 60000);
        return () =>
        {
            clearInterval(interval);
        };
    }, [credentials, document.hidden, activeTab]);

    // Save state to localStorage whenever it changes
    useEffect(() =>
    {
        localStorage.setItem('selectedID', selectedID);
        localStorage.setItem('userID', userID);
    }, [selectedID, userID]);

    // rest of your component



    useEffect(() =>
    {
        // Pulls all necessary data if never done before
        if (gridData.length !== 9 && credentials)
        {
            fetchUserMap();
            // getting the main page floor grid data
            fetchRoomsForDorms(["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"]);
            // getting the room data for uuid mapping
            fetchRoomsWithUUIDs();
        } else if (credentials)
        {
            print("Refreshing from useEffect," + refreshKey);
            fetchRoomsForOneDorm(activeTab);
            fetchRoomsWithUUIDs();
            fetchUserMap();
        }

    }, [credentials, refreshKey, activeTab]);

    // debug print function
    function print(text)
    {
        // commented console.log (text);
    }


    function fetchUserMap()
    {
        if (localStorage.getItem('jwt'))
        {
            fetch('/users/idmap', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                },
            })
                .then(res =>
                {
                    return res.json();
                })
                .then(data =>
                {
                    if (handleErrorFromTokenExpiry(data))
                    {
                        return;
                    };
                    setUserMap(data);
                })
                .catch(err =>
                {
                    // commented console.log (err);
                })
        }

    }
    function fetchRoomsWithUUIDs()
    {
        if (localStorage.getItem('jwt'))
        {
            fetch('/rooms', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`
                }
            })
                .then(res =>
                {
                    return res.json();  // Parse the response data as JSON
                })
                .then(data =>
                {
                    if (handleErrorFromTokenExpiry(data))
                    {
                        return;
                    };
                    setRooms(data);
                    // commented console.log (data);
                    if (data.error)
                    {
                        print("There was an error printing rooms");
                        setCredentials(null); // nullify the credentials if there was an error, they're probably failing
                        localStorage.removeItem('jwt');
                    }
                })
                .catch(err =>
                {
                    // commented console.log (err);
                    // commented console.log (err.error);
                })
        }
    }
    function fetchRoomsForOneDorm(dorm)
    {
        fetch(`/rooms/simple/${dorm}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
            },
        })
            .then(res => res.json())
            .then(data =>
            {
                if (handleErrorFromTokenExpiry(data))
                {
                    return;
                }
                if (!data || !data.floors || !Array.isArray(data.floors) || data.floors.length === 0)
                {
                    console.error(`Invalid data structure received for ${dorm}:`, data);
                    return;
                }

                data = splitFloorsForCaseDorm(data, roomNumbers, floorNames);

                // Ensure data.floors[0].suites exists and is an array before processing
                if (data.floors[0] && Array.isArray(data.floors[0].suites))
                {
                    data.floors.forEach(floor =>
                    {
                        if (floor && Array.isArray(floor.suites))
                        {
                            floor.suites.sort((a, b) =>
                            {
                                if (!a.rooms || !b.rooms || !Array.isArray(a.rooms) || !Array.isArray(b.rooms))
                                {
                                    return 0;
                                }
                                // Sort rooms within each suite
                                a.rooms.sort((roomA, roomB) => String(roomA?.roomNumber || '').localeCompare(String(roomB?.roomNumber || '')));
                                b.rooms.sort((roomA, roomB) => String(roomA?.roomNumber || '').localeCompare(String(roomB?.roomNumber || '')));

                                const smallestRoomNumberA = a.rooms[0]?.roomNumber || '';
                                const smallestRoomNumberB = b.rooms[0]?.roomNumber || '';
                                return String(smallestRoomNumberA).localeCompare(String(smallestRoomNumberB));
                            });
                        }
                    });
                }

                setGridData(prevGridData => prevGridData.map(item => item.dormName === dorm ? data : item));
            })
            .catch(err =>
            {
                console.error(`Error fetching rooms for ${dorm}:`, err);
            });
    }


    function fetchRoomsForDorms(dorms)
    {
        const promises = dorms.map(dorm =>
        {
            return fetch(`/rooms/simple/${dorm}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                },
            })
                .then(res => res.json())  // Parse the response data as JSON
                .then(data =>
                {
                    if (handleErrorFromTokenExpiry(data))
                    {
                        return;
                    };

                    // commented console.log ("Surely");
                    data = splitFloorsForCaseDorm(data, roomNumbers, floorNames);


                    if (Array.isArray(data.floors[0].suites))
                    {
                        data.floors.forEach(floor =>
                        {
                            if (Array.isArray(floor.suites))
                            {
                                floor.suites.sort((a, b) =>
                                {
                                    // Sort rooms within each suite
                                    a.rooms.sort((roomA, roomB) => String(roomA.roomNumber).localeCompare(String(roomB.roomNumber)));
                                    b.rooms.sort((roomA, roomB) => String(roomA.roomNumber).localeCompare(String(roomB.roomNumber)));

                                    const smallestRoomNumberA = String(a.rooms[0].roomNumber);
                                    const smallestRoomNumberB = String(b.rooms[0].roomNumber);
                                    return smallestRoomNumberA.localeCompare(smallestRoomNumberB);
                                });
                            } else
                            {
                                console.error("floor.suites is not an array:", floor.suites);
                            }
                        });
                    } else
                    {
                        console.error("data.floors[0].suites is not an array:", data.floors[0].suites);
                    }

                    return data;
                })
                .catch(err =>
                {
                    console.error(`Error fetching rooms for ${dorm}:`, err);
                });
        });

        Promise.all(promises)
            .then(dataArray =>
            {
                if (dataArray[0] && dataArray.length === 9)
                {
                    print(dataArray);
                    print("fetching roosm for dorms");
                    setGridData(dataArray);
                    // commented console.log (gridData);
                }

            })
            .catch(err =>
            {
                console.error("Error in Promise.all:", err);
            });
    }
    function splitFloorsForCaseDorm(dormData, roomNumbers, floorNames)
    {
        if (!dormData || dormData.dormName !== 'Case')
        {
            return dormData;
        }

        const newFloors = [];
        if (!dormData.floors || !Array.isArray(dormData.floors))
        {
            console.error('Invalid floors data in Case dorm:', dormData);
            return dormData;
        }

        dormData.floors.forEach((floor, index) =>
        {
            if (!floor || !Array.isArray(floor.suites))
            {
                console.error('Invalid floor data:', floor);
                return;
            }

            const firstHalfSuites = [];
            const secondHalfSuites = [];

            floor.suites.forEach(suite =>
            {
                if (!suite || !Array.isArray(suite.rooms))
                {
                    return;
                }
                const suiteHasMatchingRoom = suite.rooms.some(room => room && roomNumbers.includes(room.roomNumber));
                if (suiteHasMatchingRoom)
                {
                    firstHalfSuites.push(suite);
                } else
                {
                    secondHalfSuites.push(suite);
                }
            });

            newFloors.push({
                ...floor,
                floorNumber: floor.floorNumber,
                floorName: floorNames[floor.floorNumber * 2],
                suites: firstHalfSuites,
            });

            newFloors.push({
                ...floor,
                floorNumber: floor.floorNumber,
                floorName: floorNames[floor.floorNumber * 2 + 1],
                suites: secondHalfSuites,
            });
        });

        return {
            ...dormData,
            floors: newFloors,
        };
    }
    // fixed mapping from dorms to numbers

    const dormMapping = {
        "1": "East",
        "2": "North",
        "3": "South",
        "4": "West",
        "5": "Atwood",
        "6": "Sontag",
        "7": "Case",
        "8": "Drinkward",
        "9": "Linde"
    };


    const cellColors = {
        name: "Default",
        unbumpableRoom: "white",
        roomNumber: "#ffd6ff",
        pullMethod: "#ffbbf2",
        evenSuite: "#ffc8dd",
        oddSuite: "#ffbbf2",
        selectedUserRoom: "#3a86ff",
        currentUserRoom: "#ff006e",
    };

    const cellColors2 = {
        name: "Starburst",
        unbumpableRoom: "white",
        roomNumber: "#9e0059",
        pullMethod: "#ff7d00",
        evenSuite: "#ffbd00",
        oddSuite: "#ff5400",
        selectedUserRoom: "#ff0054",
        currentUserRoom: "#ff5c8a",
    };

    const cellColors3 = {
        name: "High contrast",
        unbumpableRoom: "white",
        roomNumber: "#9fb8ad",
        pullMethod: "#FF7B25",
        evenSuite: "#ffebc6",
        oddSuite: "#ffb100",
        selectedUserRoom: "#3a0ca3",
        currentUserRoom: "#f72585",
    };

    const modernPalette = {
        name: "Modern",
        unbumpableRoom: "white",
        roomNumber: "#81ecec",
        pullMethod: "#00cec9",
        evenSuite: "#74b9ff",
        oddSuite: "#0984e3",
        selectedUserRoom: "#8e44ad",
        currentUserRoom: "#e84393",
    };

    const pastelPalette = {
        name: "Pastel Dream",
        unbumpableRoom: "white",
        roomNumber: "#a8e6cf",
        pullMethod: "#dcedc1",
        evenSuite: "#ffd3b6",
        oddSuite: "#ffaaa5",
        selectedUserRoom: "#7209b7",
        currentUserRoom: "#f72585",
    };

    const mintChocolatePalette = {
        name: "Mint Chocolate",
        unbumpableRoom: "white",
        roomNumber: "#9de0ad",
        pullMethod: "#c2efb3",
        evenSuite: "#d2c2b0",
        oddSuite: "#e6d7c3",
        selectedUserRoom: "#38b000",
        currentUserRoom: "#ff0a54",
    };

    const oceanBreezePalette = {
        name: "Ocean Breeze",
        unbumpableRoom: "white",
        roomNumber: "#a8dadc",
        pullMethod: "#457b9d",
        evenSuite: "#e1e5f2",
        oddSuite: "#caf0f8",
        selectedUserRoom: "#0077b6",
        currentUserRoom: "#e63946",
    };

    const monochromePalette = {
        name: "Monochrome",
        unbumpableRoom: "white",
        roomNumber: "#f8f9fa",
        pullMethod: "#e9ecef",
        evenSuite: "#dee2e6",
        oddSuite: "#ced4da",
        selectedUserRoom: "#495057",
        currentUserRoom: "#212529",
    };

    const autumnPalette = {
        name: "Autumn",
        unbumpableRoom: "#3a3335",
        roomNumber: "#f0a868",
        pullMethod: "#d08c60",
        evenSuite: "#e4d6a7",
        oddSuite: "#9e9d89",
        selectedUserRoom: "#540b0e",
        currentUserRoom: "#9e2a2b",
    };

    const lavenderDreamPalette = {
        name: "Lavender Dream",
        unbumpableRoom: "white",
        roomNumber: "#e9d8fd",
        pullMethod: "#d6bcfa",
        evenSuite: "#b794f4",
        oddSuite: "#805ad5",
        selectedUserRoom: "#4c1d95",
        currentUserRoom: "#e11d48",
    };

    const cellColors4 = {
        name: "Earth Tones",
        unbumpableRoom: "white",
        roomNumber: "#faedcd",
        pullMethod: "#fefae0",
        evenSuite: "#e9edc9",
        oddSuite: "#ccd5ae",
        selectedUserRoom: "#283618",
        currentUserRoom: "#bc6c25",
    };

    // Custom palette that will be modified by user
    const customPalette = {
        name: "Custom",
        unbumpableRoom: "white",
        roomNumber: "#f8f9fa",
        pullMethod: "#e9ecef",
        evenSuite: "#dee2e6",
        oddSuite: "#ced4da",
        selectedUserRoom: "#4834d4",
        currentUserRoom: "#eb4d4b",
    };

    // Dark mode palettes
    const darkCustomPalette = {
        name: "Custom",
        unbumpableRoom: "black",
        roomNumber: "#050607",
        pullMethod: "#0f1316",
        evenSuite: "#191c21",
        oddSuite: "#252b31",
        selectedUserRoom: "#3e2aca",
        currentUserRoom: "#b41513",
    };

    const darkDefaultPalette = {
        name: "Default",
        unbumpableRoom: "black",
        roomNumber: "#280028",
        pullMethod: "#430036",
        evenSuite: "#370015",
        oddSuite: "#430036",
        selectedUserRoom: "#004bc4",
        currentUserRoom: "#ff006d",
    };

    const darkModernPalette = {
        name: "Modern",
        unbumpableRoom: "black",
        roomNumber: "#127e7e",
        pullMethod: "#30fff9",
        evenSuite: "#00448b",
        oddSuite: "#1b96f6",
        selectedUserRoom: "#9c51bb",
        currentUserRoom: "#bc1666",
    };

    const darkPastelPalette = {
        name: "Pastel Dream",
        unbumpableRoom: "black",
        roomNumber: "#185640",
        pullMethod: "#2d3e11",
        evenSuite: "#491c00",
        oddSuite: "#5a0400",
        selectedUserRoom: "#b048f6",
        currentUserRoom: "#da0867",
    };

    const darkMintChocolatePalette = {
        name: "Mint Chocolate",
        unbumpableRoom: "black",
        roomNumber: "#1f622f",
        pullMethod: "#1e4c10",
        evenSuite: "#4f3e2c",
        oddSuite: "#3c2d19",
        selectedUserRoom: "#87ff4f",
        currentUserRoom: "#f4004a",
    };

    const darkOceanBreezePalette = {
        name: "Ocean Breeze",
        unbumpableRoom: "black",
        roomNumber: "#235457",
        pullMethod: "#6197ba",
        evenSuite: "#0d111e",
        oddSuite: "#072c35",
        selectedUserRoom: "#48c0ff",
        currentUserRoom: "#c61825",
    };

    const darkMonochromePalette = {
        name: "Monochrome",
        unbumpableRoom: "black",
        roomNumber: "#050607",
        pullMethod: "#0f1316",
        evenSuite: "#191c21",
        oddSuite: "#252b31",
        selectedUserRoom: "#a8afb6",
        currentUserRoom: "#d5dade",
    };

    const darkAutumnPalette = {
        name: "Autumn",
        unbumpableRoom: "black",
        roomNumber: "#964e0f",
        pullMethod: "#9f5b2f",
        evenSuite: "#58491a",
        oddSuite: "#767561",
        selectedUserRoom: "#f4abad",
        currentUserRoom: "#d56061",
    };

    const darkLavenderDreamPalette = {
        name: "Lavender Dream",
        unbumpableRoom: "black",
        roomNumber: "#130126",
        pullMethod: "#1e0442",
        evenSuite: "#2d0a6a",
        oddSuite: "#4f29a5",
        selectedUserRoom: "#9869e2",
        currentUserRoom: "#e21e49",
    };

    const darkEarthTonesPalette = {
        name: "Earth Tones",
        unbumpableRoom: "black",
        roomNumber: "#322505",
        pullMethod: "#1f1b00",
        evenSuite: "#313511",
        oddSuite: "#48512a",
        selectedUserRoom: "#d9e7c8",
        currentUserRoom: "#da8942",
    };

    const colorPalettes = [
        customPalette,
        cellColors,
        modernPalette,
        pastelPalette,
        oceanBreezePalette,
        mintChocolatePalette,
        lavenderDreamPalette,
        autumnPalette,
        monochromePalette,
        cellColors4,
        cellColors3,
    ];

    const darkColorPalettes = [
        darkCustomPalette,
        darkDefaultPalette,
        darkModernPalette,
        darkPastelPalette,
        darkOceanBreezePalette,
        darkMintChocolatePalette,
        darkLavenderDreamPalette,
        darkAutumnPalette,
        darkMonochromePalette,
        darkEarthTonesPalette,
        darkDefaultPalette, // Just use the default dark palette for the last one
    ];

    const [isDarkMode, setIsDarkMode] = useState(() =>
    {
        const savedMode = localStorage.getItem('darkMode');
        return savedMode ? JSON.parse(savedMode) : false;
    });

    // Effect to update body class when dark mode changes
    useEffect(() =>
    {
        if (isDarkMode)
        {
            document.body.classList.add('dark-mode');
        } else
        {
            document.body.classList.remove('dark-mode');
        }
        localStorage.setItem('darkMode', JSON.stringify(isDarkMode));
    }, [isDarkMode]);

    const [selectedPalette, setSelectedPalette] = useState(() =>
    {
        const savedMode = localStorage.getItem('darkMode');
        const isDark = savedMode ? JSON.parse(savedMode) : false;

        const storedPalette = localStorage.getItem('selectedPalette');
        if (storedPalette)
        {
            const palette = JSON.parse(storedPalette);
            // Check if the stored palette matches the current mode
            if (isDark)
            {
                // If dark mode, make sure we use a dark palette
                const darkPaletteIndex = darkColorPalettes.findIndex(p => p.name === palette.name);
                return darkPaletteIndex !== -1 ? darkColorPalettes[darkPaletteIndex] : darkColorPalettes[0];
            } else
            {
                // If light mode, make sure we use a light palette
                const lightPaletteIndex = colorPalettes.findIndex(p => p.name === palette.name);
                return lightPaletteIndex !== -1 ? colorPalettes[lightPaletteIndex] : colorPalettes[0];
            }
        }

        // Default to first palette in appropriate array
        return isDark ? darkColorPalettes[0] : colorPalettes[0];
    });

    // Toggle dark mode function
    const toggleDarkMode = () =>
    {
        // Get current palette name before toggling
        const currentPaletteName = selectedPalette.name;
        const willBeDarkMode = !isDarkMode;

        // Toggle dark mode
        setIsDarkMode(willBeDarkMode);

        // Switch palette based on new mode
        if (willBeDarkMode)
        {
            // Switch to dark equivalent
            const darkPaletteIndex = darkColorPalettes.findIndex(palette => palette.name === currentPaletteName);
            if (darkPaletteIndex !== -1)
            {
                setSelectedPalette(darkColorPalettes[darkPaletteIndex]);
            } else
            {
                // Fallback to first dark palette
                setSelectedPalette(darkColorPalettes[0]);
            }
        } else
        {
            // Switch to light equivalent
            const lightPaletteIndex = colorPalettes.findIndex(palette => palette.name === currentPaletteName);
            if (lightPaletteIndex !== -1)
            {
                setSelectedPalette(colorPalettes[lightPaletteIndex]);
            } else
            {
                // Fallback to first light palette
                setSelectedPalette(colorPalettes[0]);
            }
        }
    };

    const getNameById = (id) =>
    {
        if (id === -1)
        {
            return "Frosh!!!";
        }
        // given an ID, return the First and Last name of the user
        if (id && userMap)
        {
            id = id.toString();
            if (userMap[id] === undefined)
            {
                return '';
            }
            return `${userMap[id].FirstName} ${userMap[id].LastName}`;
        }
        return "";

    };

    const handleTakeMeThere = (myLocationString, isCurrentUser = false) =>
    {
        const words = myLocationString.split(' ');
        if (words.length === 2)
        {
            setActiveTab(words[0]);
        }

        // Get the room UUID based on whether it's the current user or selected user
        const targetUserID = isCurrentUser ? userID : selectedID;
        const roomUUID = getRoomUUIDFromUserID(targetUserID);

        // Delay the scrolling until after the tab has finished switching
        setTimeout(() =>
        {
            const roomRef = roomRefs.current[roomUUID];
            if (roomRef)
            {
                // Calculate position to scroll to (element's top position - half viewport height)
                const elementRect = roomRef.getBoundingClientRect();
                const absoluteElementTop = elementRect.top + window.pageYOffset;
                const middle = absoluteElementTop - (window.innerHeight / 2);
                
                window.scrollTo({
                    top: middle,
                    behavior: 'smooth'
                });
            }
        }, 0);
    }

    const sharedData = {
        currPage,
        refreshKey,
        setRefreshKey,
        setCurrPage,
        isModalOpen,
        setIsModalOpen,
        gridData,
        setGridData,
        selectedItem,
        setSelectedItem,
        selectedOccupants,
        setSelectedOccupants,
        pullMethod,
        setPullMethod,
        showModalError,
        setShowModalError,
        onlyShowBumpableRooms,
        setOnlyShowBumpableRooms,
        userMap,
        dormMapping,
        getNameById,
        selectedRoomObject,
        setSelectedRoomObject,
        cellColors,
        rooms,
        pullError,
        setPullError,
        selectedID,
        setSelectedID,
        userID,
        setUserID,
        selectedSuiteObject,
        setSelectedSuiteObject,
        print,
        isSuiteNoteModalOpen,
        setIsSuiteNoteModalOpen,
        credentials,
        setCredentials,
        lastRefreshedTime,
        setLastRefreshedTime,
        activeTab,
        setActiveTab,
        handleErrorFromTokenExpiry,
        isFroshModalOpen,
        setIsFroshModalOpen,
        suiteDimensions,
        setSuiteDimensions,
        getRoomUUIDFromUserID,
        roomRefs,
        colorPalettes,
        darkColorPalettes,
        selectedPalette,
        setSelectedPalette,
        setIsSettingsModalOpen,
        isSettingsModalOpen,
        showFloorplans,
        setShowFloorplans,
        adminList,
        isDarkMode,
        setIsDarkMode,
        toggleDarkMode,
        isUserSettingsModalOpen,
        setIsUserSettingsModalOpen,
        myRoom,
        setMyRoom,
        handleTakeMeThere
    };

    return (
        <MyContext.Provider value={sharedData}>
            {children}
        </MyContext.Provider>
    );
};