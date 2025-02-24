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
    const roomRefs = useRef({}); // references to all the room divs
    const [showFloorplans, setShowFloorplans] = useState(false);

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
    }, [selectedID]);

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
        "9": "Linde",
        "10": "Garrett House"
    };


    const cellColors = {
        name: "Default",
        unbumpableRoom: "black",
        roomNumber: "#ffd6ff",
        pullMethod: "#ffbbf2",
        evenSuite: "#ffc8dd",
        oddSuite: "#ffbbf2",
        myRoom: "#a2d2ff",

    };
    const cellColors2 = {
        name: "Starburst",
        unbumpableRoom: "#390099",
        roomNumber: "#9e0059",
        pullMethod: "#ff7d00",
        evenSuite: "#ffbd00",
        oddSuite: "#ff5400",
        myRoom: "#ff0054",
    };
    const cellColors3 = {
        name: "High contrast",
        unbumpableRoom: "#003844",
        roomNumber: "#9fb8ad",
        pullMethod: "#FF7B25",
        evenSuite: "#ffebc6",
        oddSuite: "#ffb100",
        myRoom: "#f194b4",
    };
    const cellColors4 = {
        name: "Earth Tones",
        unbumpableRoom: "#588157",
        roomNumber: "#faedcd",
        pullMethod: "#fefae0",
        evenSuite: "#e9edc9",
        oddSuite: "#ccd5ae",
        myRoom: "#d4a373",
    };

    const colorPalettes = [
        cellColors, cellColors4, cellColors3, cellColors2
    ]

    const [selectedPalette, setSelectedPalette] = useState(() =>
    {
        const storedPalette = localStorage.getItem('selectedPalette');
        return storedPalette ? JSON.parse(storedPalette) : colorPalettes[0];
    });


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
        selectedPalette,
        setSelectedPalette,
        setIsSettingsModalOpen,
        isSettingsModalOpen,
        showFloorplans,
        setShowFloorplans,
        adminList

    };

    return (
        <MyContext.Provider value={sharedData}>
            {children}
        </MyContext.Provider>
    );
};