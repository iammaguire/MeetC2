namespace Module {
    using System;
    using System.IO;

    static class TestModule2 {
        static void Main(string[] args) {
            Console.WriteLine("Hello world!"); 
            Console.WriteLine("Hello world!"); 
            Console.WriteLine("Hello world!"); 
            StreamWriter sw = new StreamWriter("C:\\users\\meet\\desktop\\Test.txt");
            sw.WriteLine("Hello World!!");
            sw.WriteLine("From the StreamWriter class");
            sw.Close();
        } 
    }
}