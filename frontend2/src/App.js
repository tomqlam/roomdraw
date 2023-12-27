import React from 'react';
import 'bulma/css/bulma.min.css';
import BumpModal from './BumpModal';
import { useState, useEffect } from 'react';
import FloorGrid from './FloorGrid';
import Recommendations from './Recommendations';

function App() {
  const [currPage, setCurrPage] = useState('Home');
  const [showNotification, setShowNotification] = useState(false);

  const handleDeleteClick = () => {
    setShowNotification(false);
  };

  const cellColors = {
    roomNumber: "#ffd6ff",
    occupants: "#ffc8dd",
    pullMethod: "#ffbbf2",

  }
  // const tabs = ["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"]
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

  const [gridData, setGridData] = useState([
    {
      dormName: 'Atwood', description: "Description: Mixed dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
        [{ roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
        { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' },],
      ]
    },
    {
      dormName: 'East', description: "Description: boring dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
        [{ roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
        { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' },],
      ]
    },
    {
      dormName: 'West', description: "Description: less fun dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
      ]
    },
    {
      dormName: 'South', description: "Description: super loud", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
      ]
    },
    {
      dormName: 'North', description: "Description: a cool dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
        [{ roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
        { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' }],
      ]
    },
    {
      dormName: 'Sontag', description: "Description: a cool dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },
        { roomNumber: 105, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 107, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' }],
      ]
    },
    {
      dormName: 'Case', description: "Description: a cool dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
      ]
    },
    {
      dormName: 'Linde', description: "Description: a cool dorm", floors: [
        [{ roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
        { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },],
      ]
    },
  ])
  // const [gridData, setGridData] = useState([[[
  //   { header: "atwood floor 1" },
  //   { roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
  //   { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },
  //   { roomNumber: 105, notes: 'in dorm 143 pull', occupant1: 'carlos sanchez', occupant2: '', occupant3: '' },
  //   { roomNumber: 107, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: 'frosh' },
  //   { roomNumber: 109, notes: 'senior 74', occupant1: 'tejas hegde', occupant2: '', occupant3: '' },
  //   { roomNumber: 111, notes: 'senior 74 pull', occupant1: 'chris morales', occupant2: '', occupant3: '' },
  //   { roomNumber: 113, notes: 'sophomore 32', occupant1: 'jared carreno', occupant2: 'terence chen', occupant3: '' },
  //   { roomNumber: 115, notes: '', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 117, notes: 'in-dorm 96', occupant1: 'dylan camacho', occupant2: '', occupant3: '' },
  //   { roomNumber: 119, notes: 'in-dorm 96 pull', occupant1: 'josh gk', occupant2: '', occupant3: '' },
  //   { roomNumber: 121, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: 'frosh' },
  //   { roomNumber: 123, notes: 'in-dorm 35', occupant1: 'arjun asija', occupant2: '', occupant3: '' },
  //   { roomNumber: 125, notes: 'in-dorm pull', occupant1: 'alec candidato', occupant2: '', occupant3: '' },
  //   { roomNumber: 127, notes: 'lock pull', occupant1: 'jackson king', occupant2: 'josiah garan', occupant3: '' },
  //   { roomNumber: 100, notes: 'senior 95 lock pull', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 102, notes: 'preplaced', occupant1: 'tresselle gatutha', occupant2: '', occupant3: '' },
  //   { roomNumber: 104, notes: 'proctor', occupant1: 'marina ring', occupant2: '', occupant3: '' },
  //   { roomNumber: 106, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: 'frosh' },
  //   { roomNumber: 108, notes: 'senior 95 pull', occupant1: 'alex pedroza', occupant2: '', occupant3: '' },
  //   { roomNumber: 110, notes: 'senior 95', occupant1: 'sydney riley', occupant2: '', occupant3: '' },
  //   { roomNumber: 112, notes: 'efficiency jr32', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 114, notes: 'efficiency', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 116, notes: 'senior 11 pull', occupant1: 'fred bolarinwa', occupant2: '', occupant3: '' },
  //   { roomNumber: 118, notes: 'senior 11', occupant1: 'elijah adamson', occupant2: '', occupant3: '' },
  //   { roomNumber: 120, notes: 'frosh', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
  //   { roomNumber: 122, notes: 'mentor', occupant1: 'nile phillips', occupant2: '', occupant3: '' },
  //   { roomNumber: 124, notes: 'mentor pull', occupant1: 'lucas grandison', occupant2: '', occupant3: '' },
  //   { roomNumber: 126, notes: 'lock pull', occupant1: 'jeremy tan', occupant2: 'tyler headley', occupant3: '' },
  // ],
  // [
  //   { header: "atwood floor 2" },
  // { roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
  // { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' },
  //   { roomNumber: 205, notes: 'in dorm 115 lock pull', occupant1: 'marika ragnartz', occupant2: '', occupant3: '' },
  //   { roomNumber: 207, notes: 'mentor', occupant1: 'abigail samson', occupant2: 'kaitlyn chen', occupant3: 'delaney pratt' },
  //   { roomNumber: 209, notes: 'senior 56 atwood', occupant1: 'vani sachdev', occupant2: '', occupant3: '' },
  //   { roomNumber: 211, notes: 'senior 56 pull', occupant1: 'cheyenne foo', occupant2: '', occupant3: '' },
  //   { roomNumber: 213, notes: 'sophomore 23', occupant1: 'adam tang', occupant2: 'steven tran', occupant3: '' },
  //   { roomNumber: 215, notes: 'sophomore 4', occupant1: 'ben colbeck', occupant2: 'manan mendi', occupant3: '' },
  //   { roomNumber: 217, notes: 'in-dorm 27', occupant1: 'nilay pangrekar', occupant2: '', occupant3: '' },
  //   { roomNumber: 219, notes: 'senior 27 pull', occupant1: 'eli schwarz', occupant2: '', occupant3: '' },
  //   { roomNumber: 221, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
  // ]], [[
  //   { header: "sontag floor 1" },
  //   { roomNumber: 101, notes: 'lock pull', occupant1: 'troy kaufman', occupant2: 'stuart kerr', occupant3: '' },
  //   { roomNumber: 103, notes: 'in dorm 143', occupant1: 'javier perez', occupant2: '', occupant3: '' },
  //   { roomNumber: 105, notes: 'in dorm 143 pull', occupant1: 'carlos sanchez', occupant2: '', occupant3: '' },
  //   { roomNumber: 107, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: 'frosh' },
  //   { roomNumber: 109, notes: 'senior 74', occupant1: 'tejas hegde', occupant2: '', occupant3: '' },
  //   { roomNumber: 111, notes: 'senior 74 pull', occupant1: 'chris morales', occupant2: '', occupant3: '' },
  //   { roomNumber: 113, notes: 'sophomore 32', occupant1: 'jared carreno', occupant2: 'terence chen', occupant3: '' },
  //   { roomNumber: 115, notes: '', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 117, notes: 'in-dorm 96', occupant1: 'dylan camacho', occupant2: '', occupant3: '' },
  //   { roomNumber: 119, notes: 'in-dorm 96 pull', occupant1: 'josh gk', occupant2: '', occupant3: '' },
  //   { roomNumber: 121, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: 'frosh' },
  //   { roomNumber: 123, notes: 'in-dorm 35', occupant1: 'arjun asija', occupant2: '', occupant3: '' },
  //   { roomNumber: 125, notes: 'in-dorm pull', occupant1: 'alec candidato', occupant2: '', occupant3: '' },
  //   { roomNumber: 127, notes: 'lock pull', occupant1: 'jackson king', occupant2: 'josiah garan', occupant3: '' },
  //   { roomNumber: 100, notes: 'senior 95 lock pull', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 102, notes: 'preplaced', occupant1: 'tresselle gatutha', occupant2: '', occupant3: '' },
  //   { roomNumber: 104, notes: 'proctor', occupant1: 'marina ring', occupant2: '', occupant3: '' },
  //   { roomNumber: 106, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: 'frosh' },
  //   { roomNumber: 108, notes: 'senior 95 pull', occupant1: 'alex pedroza', occupant2: '', occupant3: '' },
  //   { roomNumber: 110, notes: 'senior 95', occupant1: 'sydney riley', occupant2: '', occupant3: '' },
  //   { roomNumber: 112, notes: 'efficiency jr32', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 114, notes: 'efficiency', occupant1: '', occupant2: '', occupant3: '' },
  //   { roomNumber: 116, notes: 'senior 11 pull', occupant1: 'fred bolarinwa', occupant2: '', occupant3: '' },
  //   { roomNumber: 118, notes: 'senior 11', occupant1: 'elijah adamson', occupant2: '', occupant3: '' },
  //   { roomNumber: 120, notes: 'frosh', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
  //   { roomNumber: 122, notes: 'mentor', occupant1: 'nile phillips', occupant2: '', occupant3: '' },
  //   { roomNumber: 124, notes: 'mentor pull', occupant1: 'lucas grandison', occupant2: '', occupant3: '' },
  //   { roomNumber: 126, notes: 'lock pull', occupant1: 'jeremy tan', occupant2: 'tyler headley', occupant3: '' },
  // ],
  // [
  //   { header: "sontag floor 2" },
  //   { roomNumber: 201, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
  //   { roomNumber: 203, notes: 'in dorm 115', occupant1: 'svetlana altshuler', occupant2: '', occupant3: '' },
  //   { roomNumber: 205, notes: 'in dorm 115 lock pull', occupant1: 'marika ragnartz', occupant2: '', occupant3: '' },
  //   { roomNumber: 207, notes: 'mentor', occupant1: 'abigail samson', occupant2: 'kaitlyn chen', occupant3: 'delaney pratt' },
  //   { roomNumber: 209, notes: 'senior 56 atwood', occupant1: 'vani sachdev', occupant2: '', occupant3: '' },
  //   { roomNumber: 211, notes: 'senior 56 pull', occupant1: 'cheyenne foo', occupant2: '', occupant3: '' },
  //   { roomNumber: 213, notes: 'sophomore 23', occupant1: 'adam tang', occupant2: 'steven tran', occupant3: '' },
  //   { roomNumber: 215, notes: 'sophomore 4', occupant1: 'ben colbeck', occupant2: 'manan mendi', occupant3: '' },
  //   { roomNumber: 217, notes: 'in-dorm 27', occupant1: 'nilay pangrekar', occupant2: '', occupant3: '' },
  //   { roomNumber: 219, notes: 'senior 27 pull', occupant1: 'eli schwarz', occupant2: '', occupant3: '' },
  //   { roomNumber: 221, notes: '', occupant1: 'frosh', occupant2: 'frosh', occupant3: '' },
  // ]]
  // ]);
  const [selectedName, setSelectedName] = useState('Becca Verghese');
  const [activeTab, setActiveTab] = useState('Atwood');
  console.log(gridData);

  const handleNameChange = (event) => {
    setSelectedName(event.target.value);
  };
  const handleTabClick = (tab) => {
    setActiveTab(tab);
  };

  function updateGridData(roomData) {
    setShowNotification(true);
    setGridData((prevGridData) => {
      const updatedGridData = prevGridData.map((dorm) => {
        const updatedFloors = dorm.floors.map((floor) => {
          return floor.map((room) => {
            if (room.roomNumber === roomData.roomNumber) {
              return { ...room, ...roomData };
            }
            return room;
          });
        });
        return { ...dorm, floors: updatedFloors };
      });
      return updatedGridData;
    });
  }

  function getDrawNumberByName(name) {
    const foundItem = drawNumbers.find((item) => item.name === name);
    return foundItem ? foundItem.drawNumber : null;
  }

  const getRoom = (name) => {
    // given a name, get the room and dorm that they are in 
    // handle case with no room yet
    return "[TODO Placeholder for " + name + "'s room number]";
  }



  // const updateGridData = (roomData) => {
  //   const updatedGridData = [...gridData];
  //   const roomIndex = updatedGridData.findIndex((item) => item.roomNumber === roomData.roomNumber);
  //   if (roomIndex !== -1) {
  //     updatedGridData[roomIndex] = roomData;
  //     setGridData(updatedGridData);
  //   }
  // };

  return (
    <div>

      <nav class="navbar" role="navigation" aria-label="main navigation">
        <div class="navbar-brand">
          <a class="navbar-item" href="https://ibb.co/c3D21bJ"><img src="https://i.ibb.co/SyRVPQN/Screenshot-2023-12-26-at-10-14-31-PM.png" alt="Screenshot-2023-12-26-at-10-14-31-PM" border="0" /></a>

          <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="navbarBasicExample">
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
          </a>
        </div>

        <div id="navbarBasicExample" class="navbar-menu">
          <div class="navbar-start">
            <a class="navbar-item" onClick={() => setCurrPage("Home")}>
              Home
            </a>

            <a class="navbar-item" onClick={() => setCurrPage("Recommendations")}>
              Recommendations
            </a>

            <div class="navbar-item has-dropdown is-hoverable">
              <a class="navbar-link">
                More
              </a>

              <div class="navbar-dropdown">
                <a class="navbar-item">
                  About
                </a>
                <a class="navbar-item">
                  Jobs
                </a>
                <a class="navbar-item">
                  Contact
                </a>

                <a class="navbar-item">
                  Report an issue
                </a>
              </div>
            </div>
          </div>

          <div class="navbar-end">
            <div class="navbar-item">
              <div class="buttons">
                <a class="button is-primary">
                  <strong>Sign up</strong>
                </a>
                <a class="button is-light">
                  Log in
                </a>
              </div>
            </div>
          </div>
        </div>
      </nav>
      {showNotification && (<div class="notification is-primary m-2">
        <button onClick={handleDeleteClick} class="delete "></button>
        Your room status has been updated. Please check that everything is still the way you'd like it to be!
      </div>)}
      <section class="section">
        <div style={{ textAlign: 'center' }}>

          <h1 className="title">Welcome back, <strong>{selectedName}</strong>. <br /> </h1>
          <h2 className="subtitle">You are <strong>{getDrawNumberByName(selectedName)}</strong>. You are currently in <strong>{getRoom(selectedName)}</strong>. <br />Click on any room you'd like to change!</h2>

          <div style={{ display: 'flex', justifyContent: 'center' }}>
            <select className="select" onChange={handleNameChange}>
              <option value="">This isn't me</option>
              {drawNumbers.map((item, index) => (
                <option key={index} value={item.name}>
                  {item.name}
                </option>
              ))}
            </select>
          </div>
        </div>

      </section>

      { currPage == "Home" && <section class="section">
        <div className="tabs is-centered">
          <ul>
            {gridData.map((dorm) => (
              <li
                key={dorm.dormName}
                className={activeTab === dorm.dormName ? 'is-active' : ''}
                onClick={() => handleTabClick(dorm.dormName)}
              >
                <a>{dorm.dormName}</a>
              </li>
            ))}
          </ul>
        </div>

        {/* Left column is room draw, right side is tips */}
        <div class="columns">
          <div class="column">
            {gridData.map((dorm) => (

              <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
                <h1 class="title has-text-centered">{dorm.dormName}</h1>

                {dorm.floors.map((floor, floorIndex) => (
                  <div style={{ paddingBottom: 20 }} className="container" key={floorIndex}>

                    <h2 class="subtitle has-text-centered">Floor {floorIndex + 1}</h2>
                    <ul>
                      <FloorGrid cellColors={cellColors} gridData={floor} dropdownOptions={drawNumbers.map((number) => number.name)} updateGridData={updateGridData} />
                    </ul>
                  </div>
                ))}
              </div>
            ))}
          </div>
          <div class="column is-one-quarter">
            {gridData.map((dorm) => (

              <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
                <p>{dorm.description}</p>
              </div>
            ))}
          </div>
        </div>

        {/* {(activeTab === 'Recommendations' && (<Recommendations />))} */}

      </section>}
      { currPage == "Recommendations" && <section class="section">
        <Recommendations />
        </section>}

      
      <footer class="footer">
        <div class="content has-text-centered">
          <p>
            <strong>Digital Draw</strong> by Serena Mao & Tom Lam. Email smao@g.hmc.edu or tlam@g.hmc.edu with questions.
          </p>
        </div>
      </footer>


    </div>

  );
}

export default App;