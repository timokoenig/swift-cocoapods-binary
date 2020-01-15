# Swift Cocoapods Binary

A simple cli tool to build binary frameworks from your cocoapods frameworks with the help of CocoaPods/Rome.

## Prerequisites

RubyGems is required to make this cli tool work.

It will automatically try to install CocoaPods and CocoaPods/Rome when you haven't already.

## Usage

```
  -ios string
    	The iOS version; Default: 11.0
  -pod string
    	Name of the pod that you would like to have as a binary
  -source string
    	The podspec sources (separate with comma if more than one is needed)
  -version string
    	The pod version
```

Example
```
./swift-cocoapods-binary -pod Alamofire -version 4.9.1 
```

The generation might take a while, depending on the size of the source code and the amount of dependencies. The result will be a zip archive in the current directory containing all required binary frameworks.

## Read more

[https://github.com/CocoaPods/CocoaPods](https://github.com/CocoaPods/CocoaPods)

[https://github.com/CocoaPods/Rome](https://github.com/CocoaPods/Rome)
