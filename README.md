# multisnake

This is a multiplayer variant of snake game, written in Go on top of the SDL library.

Features

* [x] configurable settings (see below)
* [x] keyboard support
* [ ] controller support 
* [x] supports obstacles on the game error

## Getting started:

You need to have golang compiler and SDK to be installed. 
Also following dependencies are required (Ubuntu):

```shell
$ sudo apt-get install libsdl2{,-image,-mixer,-ttf,-gfx}-dev
```

Checkout this project and run 
```
$ go build
```

and the compiled executable `./multisnake config.json`.


## Configuration

The game can be configured by a JSON file, which is backed up the structure:

```json
{
  "Width": 800,     // the width in pixels of the game area
  "Height": 800,    // the height in pixels of the game area
  "Players": [      //settings for each player
    {
      "Name": "me",  //player name
      "Color": 620691711, //color of the snake
      "KeyUp": "UP",      //key for going up
      "KeyDown": "DOWN",  //key for going down
      "KeyLeft": "LEFT",  //key for going left
      "KeyRight": "RIGHT" //key for going right
    },
    {
      "Name": "me",
      "Color": 6000017,
      "KeyUp": "w",
      "KeyDown": "s",
      "KeyLeft": "a",
      "KeyRight": "d"
    }
  ],
  
  //if a snake leaves the arena, it will appear on the opposite side
  "CycleBorder": true,    
  "Food": 5, //number of foods on the arena

  //Size of a cell in the raster. Used to make snake and objects bigger
  "FieldSize": 10, 
  //Color of the food
  "FoodColor": 4278190335,
  //Background color of the game arena
  "BackgroundColor": 3435973887,
  //List of b/w PNG images, which describes the obstacle in the game arena
  //The of the images need to be Width/FieldSize x Height/FieldSize. 
  //In this example the PNG is 80x80 pixels. 
  //Obstacles are randomly chosen at game start. 
  "Obstacles": [
    "obstacles.png"
  ],
  "ObstaclesColor": 0 //Color of the obstacles
}
```