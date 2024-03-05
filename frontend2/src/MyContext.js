import React, { createContext, useState, useEffect } from 'react';


export const MyContext = createContext();

export const MyContextProvider = ({ children }) => {
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

    // Initialize active tab state from localStorage or default to 'Atwood'
    const [activeTab, setActiveTab] = useState(() => {
        const savedTab = localStorage.getItem('activeTab');
        return savedTab !== null ? savedTab : 'Atwood';
    });

    const [selectedID, setSelectedID] = useState(() => {
        const selectedID = localStorage.getItem('selectedID');
        return selectedID !== null ? selectedID : '8'; //TODO 
    });

    const handleErrorFromTokenExpiry = (data) => {
        if (data.error === "Invalid token") {
            setCredentials(null);
            localStorage.removeItem('jwt');
            return true;
        }
        return false;
    }

    useEffect(() => {
        const interval = setInterval(() => {
            if (credentials && !document.hidden) {
                setRefreshKey(prevKey => prevKey + 1);
                setLastRefreshedTime(new Date());
                console.log("refreshed ONE DORM");
            }
        }, 60000);
        return () => {
            clearInterval(interval);
        };
    }, [credentials, document.hidden, activeTab]);

    // Save state to localStorage whenever it changes
    useEffect(() => {
        localStorage.setItem('selectedID', selectedID);
    }, [selectedID]);

    // rest of your component



    useEffect(() => {
        // Pulls all necessary data if never done before
        if (gridData.length !== 9 && credentials) {
            fetchUserMap();
            // getting the main page floor grid data
            fetchRoomsForDorms(["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"]);
            // getting the room data for uuid mapping
            fetchRoomsWithUUIDs();
        } else if (credentials) {
            fetchRoomsForOneDorm(activeTab);
            fetchRoomsWithUUIDs();
            fetchUserMap();
        }

    }, [credentials, refreshKey, activeTab]);

    // debug print function
    function print(text) {
        console.log(text);
    }


    function fetchUserMap() {
        if (localStorage.getItem('jwt')) {
            fetch('/users/idmap', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                },
            })
                .then(res => {
                    return res.json();
                })
                .then(data => {
                    if (handleErrorFromTokenExpiry(data)) {
                        return;
                    };
                    setUserMap(data);
                })
                .catch(err => {
                    console.log(err);
                })
        }

    }
    function fetchRoomsWithUUIDs() {
        if (localStorage.getItem('jwt')) {
            fetch('/rooms', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`
                }
            })
                .then(res => {
                    return res.json();  // Parse the response data as JSON
                })
                .then(data => {
                    if (handleErrorFromTokenExpiry(data)) {
                        return;
                    };
                    setRooms(data);
                    console.log(data);
                    if (data.error) {
                        print("There was an error printing rooms");
                        setCredentials(null); // nullify the credentials if there was an error, they're probably failing
                        localStorage.removeItem('jwt');
                    }
                })
                .catch(err => {
                    console.log(err);
                    console.log(err.error);
                })
        }
    }
    function fetchRoomsForOneDorm(dorm) {
        console.log("fetching one dorm");
        fetch(`/rooms/simple/${dorm}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
            },
        })
            .then(res => res.json())  // Parse the response data as JSON
            .then(data => {
                if (handleErrorFromTokenExpiry(data)) {
                    return;
                };
                console.log("reached here");
                console.log(data.floors[0].suites);


                if (Array.isArray(data.floors[0].suites)) {
                    data.floors.forEach(floor => {
                        if (Array.isArray(floor.suites)) {
                            floor.suites.sort((a, b) => {
                                // Sort rooms within each suite
                                a.rooms.sort((roomA, roomB) => String(roomA.roomNumber).localeCompare(String(roomB.roomNumber)));
                                b.rooms.sort((roomA, roomB) => String(roomA.roomNumber).localeCompare(String(roomB.roomNumber)));

                                const smallestRoomNumberA = String(a.rooms[0].roomNumber);
                                const smallestRoomNumberB = String(b.rooms[0].roomNumber);
                                return smallestRoomNumberA.localeCompare(smallestRoomNumberB);
                            });
                        } else {
                            console.error("floor.suites is not an array:", floor.suites);
                        }
                    });
                } else {
                    console.error("data.floors[0].suites is not an array:", data.floors[0].suites);
                }

                console.log(data);
                setGridData(prevGridData => prevGridData.map(item => item.dormName === dorm ? data : item));
            })
            .catch(err => {
                console.error(`Error fetching rooms for ${dorm}:`, err);
            });
    }


    function fetchRoomsForDorms(dorms) {
        const promises = dorms.map(dorm => {
            return fetch(`/rooms/simple/${dorm}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                },
            })
                .then(res => res.json())  // Parse the response data as JSON
                .then(data => {
                    if (handleErrorFromTokenExpiry(data)) {
                        return;
                    };
                    return data;
                })
                .catch(err => {
                    console.error(`Error fetching rooms for ${dorm}:`, err);
                });
        });

        Promise.all(promises)
            .then(dataArray => {
                if (dataArray[0] && dataArray.length === 9) {
                    print(dataArray);
                    print("fetching roosm for dorms");
                    setGridData(dataArray);
                    console.log(gridData);
                }

            })
            .catch(err => {
                console.error("Error in Promise.all:", err);
            });
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
        unbumpableRoom: "black",
        roomNumber: "#ffd6ff",
        pullMethod: "#ffbbf2",
        evenSuite: "#ffc8dd",
        oddSuite: "#ffbbf2",
        myRoom: "#a2d2ff",

    };

    const getNameById = (id) => {
        if (id === -1) {
            return "Frosh!!!";
        }
        // given an ID, return the First and Last name of the user
        if (id && userMap) {
            id = id.toString();
            if (userMap[id] === undefined) {
                return 'Empty';
            }
            return `${userMap[id].FirstName} ${userMap[id].LastName}`;
        }
        return "Empty";

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
        setIsFroshModalOpen
    };

    return (
        <MyContext.Provider value={sharedData}>
            {children}
        </MyContext.Provider>
    );
};