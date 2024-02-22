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
    const [selectedID, setSelectedID] = useState(() => {
        const selectedID = localStorage.getItem('selectedID');
        return selectedID !== null ? selectedID : '8'; //TODO 
      });
    const [isSuiteNoteModalOpen, setIsSuiteNoteModalOpen] = useState(false); // If suite note modal 


  // Save state to localStorage whenever it changes
  useEffect(() => {
    localStorage.setItem('selectedID', selectedID);
  }, [selectedID]);

  // rest of your component



    useEffect(() => {
        // Pulls all necessary data
        fetchUserMap();
        // getting the main page floor grid data
        fetchRoomsForDorms(["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"]);
        // getting the room data for uuid mapping
        fetchRoomsWithUUIDs();

        // const timer = setTimeout(() => {

        // }, 0);  // Delay of 1 second

        // Clean up function
        // return () => clearTimeout(timer);
    }, [refreshKey]);

    // debug print function
    function print(text) {
        console.log(text);
    }


    function fetchUserMap() {
        fetch('/users/idmap')
            .then(res => {
                return res.json();
            })
            .then(data => {

                setUserMap(data);
            })
            .catch(err => {
                console.log(err);
            })
    }
    function fetchRoomsWithUUIDs() {
        fetch('/rooms')
            .then(res => {
                return res.json();  // Parse the response data as JSON
            })
            .then(data => {
                setRooms(data);
            })
            .catch(err => {
                console.log(err);
            })
    }

    function fetchRoomsForDorms(dorms) {
        const promises = dorms.map(dorm => {
            return fetch(`/rooms/simple/${dorm}`)
                .then(res => res.json())  // Parse the response data as JSON
                .catch(err => {
                    console.error(`Error fetching rooms for ${dorm}:`, err);
                });
        });

        Promise.all(promises)
            .then(dataArray => {
                setGridData(dataArray);
                console.log(gridData);
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
        setIsSuiteNoteModalOpen
    };

    return (
        <MyContext.Provider value={sharedData}>
            {children}
        </MyContext.Provider>
    );
};