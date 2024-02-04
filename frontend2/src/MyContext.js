import React, { createContext, useState, useEffect } from 'react';
import usersJson from './mock_data/users.json';
import usersMap from './mock_data/users_map.json';
import atwoodJson from './mock_data/atwood.json';
import caseJson from './mock_data/case.json';
import drinkwardJson from './mock_data/drinkward.json';
import eastJson from './mock_data/east.json';
import lindeJson from './mock_data/linde.json';
import northJson from './mock_data/north.json';
import sontagJson from './mock_data/sontag.json';
import southJson from './mock_data/south.json';
import westJson from './mock_data/west.json';


export const MyContext = createContext();

export const MyContextProvider = ({ children }) => {
    const [currPage, setCurrPage] = useState('Home');
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [rooms, setRooms] = useState([]);
    //const [closeModalAction, setCloseModalAction] = useState(console.log("Hello"));
    const [selectedItem, setSelectedItem] = useState(null);
    const [selectedOccupants, setSelectedOccupants] = useState(['', '', '', '']);
    const [pullMethod, setPullMethod] = useState('Select a pull method');
    const [showModalError, setShowModalError] = useState(false);
    const [onlyShowBumpableRooms, setOnlyShowBumpableRooms] = useState(false);
    const [gridData, setGridData] = useState([]);
    const [users, setUsers] = useState(usersJson);
    const [userMap, setUserMap] = useState(null);
    const [selectedRoomObject, setSelectedRoomObject] = useState(null);
    const [selectedSuiteObject, setSelectedSuiteObject] = useState(null);
    const [refreshKey, setRefreshKey] = useState(0);
    const [pullError, setPullError] = useState("There was an unknown error. Please try again.");
    const [selectedID, setSelectedID] = useState(8);


    useEffect(() => {
        const timer = setTimeout(() => {
            fetch('/users/idmap')
                .then(res => {
                    return res.json();  // Parse the response data as JSON
                })
                .then(data => {

                    setUserMap(data);
                })
                .catch(err => {
                    console.log(err);
                })
            // getting the main page floor grid data
            fetchRoomsForDorms(["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"]);
            // getting the room data for uuid mapping
            fetchRoomsWithUUIDs();
        }, 500);  // Delay of 1 second
    
        // Clean up function
        return () => clearTimeout(timer);
    }, [refreshKey]);

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
                console.log(dataArray);  // Array of data from all fetch operations
                setGridData(dataArray);
                console.log(gridData);
                // Do something with dataArray here
            })
            .catch(err => {
                console.error("Error in Promise.all:", err);
            });
    }


    // const [gridData, setGridData] = useState([
    //     {
    //         dormName: 'Atwood', description: "Description: Mixed dorm", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //             [{ roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
    //             { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' },],
    //         ]
    //     },
    //     {
    //         dormName: 'East', description: "Description: boring dorm", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //             [{ roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
    //             { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' },],
    //         ]
    //     },
    //     {
    //         dormName: 'West', description: "Description: less fun dorm", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //         ]
    //     },
    //     {
    //         dormName: 'South', description: "Description: super loud", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //         ]
    //     },
    //     {
    //         dormName: 'North', description: "Description: a cool dorm", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //             [{ roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
    //             { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' }],
    //         ]
    //     },
    //     {
    //         dormName: 'Sontag', description: "Description: a cool dorm", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },
    //             { roomNumber: 105, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 107, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' }],
    //         ]
    //     },
    //     {
    //         dormName: 'Case', description: "Description: a cool dorm", imageLinks: ["https://i.ibb.co/sQKkx4G/case1.png","https://i.ibb.co/kQkMtvW/case2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //         ]
    //     },
    //     {
    //         dormName: 'Linde', description: "Description: a cool dorm", imageLinks: ["https://i.ibb.co/ZJX43g2/atwood1.png","https://i.ibb.co/DtHkDjL/atwood2.png"], floors: [
    //             [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
    //             { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
    //         ]
    //     },
    // ])
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
        occupants: "#ffc8dd",
        pullMethod: "#ffbbf2",
        evenSuite: "#ffc8dd",
        oddSuite: "#ffbbf2",
    
      };

    //   const [data, setData] = useState('Initial data');
    //   const [count, setCount] = useState(0);
    //   const [isLoggedIn, setIsLoggedIn] = useState(false);
    const getNameById = (id) => {  
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
        // data,
        // setData,
        // count,
        // setCount,
        // isLoggedIn,
        // setIsLoggedIn,
    };

    return (
        <MyContext.Provider value={sharedData}>
            {children}
        </MyContext.Provider>
    );
};