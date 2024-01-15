import React, { createContext, useState } from 'react';
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
    //const [closeModalAction, setCloseModalAction] = useState(console.log("Hello"));
    const [selectedItem, setSelectedItem] = useState(null);
    const [selectedOccupants, setSelectedOccupants] = useState(['', '', '', '']);
    const [pullMethod, setPullMethod] = useState('');
    const [showModalError, setShowModalError] = useState(false);
    const [onlyShowBumpableRooms, setOnlyShowBumpableRooms] = useState(false);
    const [gridData, setGridData] = useState([atwoodJson, eastJson, drinkwardJson, lindeJson, northJson, southJson, sontagJson, westJson, caseJson]);
    const [users, setUsers] = useState(usersJson);
    const [userMap, setUserMap] = useState(usersMap);
    const [selectedRoomObject, setSelectedRoomObject] = useState(null);

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
    const drawNumbers = [
        { name: 'Kai Rajesh', drawNumber: 'Senior 2 East' },
        { name: 'Andres Rivas', drawNumber: 'Senior 3 Atwood' },
        { name: 'Mehek Mehra', drawNumber: 'Senior 4 Atwood' },
        { name: 'Julia Du', drawNumber: 'Senior 5 South' },
        { name: 'James Nicholson', drawNumber: 'Senior 6 East' },
        { name: 'Sophie Bekerman', drawNumber: 'Senior 7 Linde' },
        { name: 'Amy Liu', drawNumber: 'Senior 8 Atwood' },
        { name: 'Becca Verghese', drawNumber: 'Senior 9 Drinkward' },
        { name: 'Luke Stemple', drawNumber: 'Senior 10 East' },
        { name: 'Elijah Adamson', drawNumber: 'Senior 11 Linde' },
        { name: 'Toby Anderson', drawNumber: 'Senior 12 Sontag' },
        { name: 'Helen Chen', drawNumber: 'Senior 13 Atwood' },
        { name: 'Kevin Box', drawNumber: 'Senior 14 North' },
        { name: 'Tanvi Krishnan', drawNumber: 'Senior 15 Sontag' },
        { name: 'Eli Pregerson', drawNumber: 'Senior 16 Linde' },
        { name: 'Kaeshav Danesh', drawNumber: 'Senior 17 Sontag' },
      ];
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

    //   const [data, setData] = useState('Initial data');
    //   const [count, setCount] = useState(0);
    //   const [isLoggedIn, setIsLoggedIn] = useState(false);
    const getNameById = (id) => {  
        if (id) {
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
        drawNumbers,
        onlyShowBumpableRooms,
        setOnlyShowBumpableRooms,   
        userMap,
        dormMapping,
        getNameById,
        selectedRoomObject,
        setSelectedRoomObject,
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