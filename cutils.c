#include <string.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <ctype.h>
#include "cutils.h"

/* This metod is delegated to init the fsize in input with the byte lenght of
 * the given file */
void set_file_size(FILE *fp, int *fsize) {
  fseek(fp, 0, SEEK_END);
  *fsize = ftell(fp);
  rewind(fp);
  return;
}

long get_file_size(char *filename) {
  long fsize = 0;
  FILE *fp;

  fp = fopen(filename, "r");
  if (fp) {
    fseek(fp, 0, SEEK_END);
    fsize = ftell(fp);
    fclose(fp);
  }
  return fsize;
}

/* Read the file and return the content */
char *read_content(const char *filename) {
  char *fcontent = NULL;
  int fsize = 0;
  FILE *fp;
  fp = fopen(filename, "r");
  if (fp) {
    set_file_size(fp, &fsize);
    //printf("File size: %d bytes\n", fsize);
    fcontent = (char *)malloc(sizeof(char) * fsize);
    fread(fcontent, 1, fsize, fp);
    fclose(fp);
  }
  return fcontent;
}

/* Read the file and return the content */
void read_content_no_alloc(char *filename, char *out) {
  int fsize = 0;
  FILE *fp;
  fp = fopen(filename, "r");
  if (fp) {
    set_file_size(fp, &fsize);
    fread(out, 1, fsize, fp);
    fclose(fp);
  }
  free(filename);
  return;
}

// Lower the case of the string in input
void lowerize_string(char *data) {
  int i = 0; // Don't break c99 :/i
  for (; data[i]; i++) {
    data[i] = tolower(data[i]);
  }
}

int verify_presence_data_insensitive(char *data, char *to_find) {
  int ret = -1;
  /* Manage None input */
  if (data != NULL && to_find != NULL) {
    // printf("Data: %s | ToFind: %s\n",data,to_find);
    // lowering the line
    lowerize_string(data);
    /* Finding string ... */
    if (strstr(data, to_find) != NULL) {
      // printf("String %s found ..\n", to_find);
      ret = 1;
    } else {
      // printf("String %s NOT found ..\n", to_find);
      ret = 0;
    }
  }

  free(data);
  free(to_find);
  return ret;
}

int verify_presence_data(const char *data, const char *to_find) {
  /* Manage None input */
  if (data != NULL && to_find != NULL) {
    /* Finding string ... */
    if (strstr(data, to_find) != NULL) {
      // printf("String %s found ..\n",to_find);
      return 1;
    }
    // printf("String %s NOT found ..\n",to_find);
    return 0;
  } else
    // Error - input
    return -1;
}

int verify_presence_filename(const char *filename, const char *to_find) {

  /* Manage None input */
  if (filename != NULL && to_find != NULL) {

    const char *content;
    content = (char *)read_content(filename);

    /* Finding string ... */
    return verify_presence_data(content, to_find);
  }

  printf("Seems that the filename or the string to find is NULL :/ ...");
  return -1;
}

// C helper functions:

static char **makeCharArray(int size) { return calloc(sizeof(char *), size); }

static void setArrayString(char **a, char *s, int n) { a[n] = s; }


size_t levenshtein_n(const char *a, const size_t length, const char *b,
                     const size_t bLength) {
  size_t index = 0;
  size_t bIndex = 0;
  size_t bDistance;
  size_t result;

  // Shortcut optimizations / degenerate cases.
  if (a == b || strcmp(a, b) == 0) {
    return 0;
  }

  if (length == 0) {
    return bLength;
  }

  if (bLength == 0) {
    return length;
  }

  size_t *cache = calloc(length, sizeof(size_t));

  // initialize the vector.
  while (index < length) {
    cache[index] = index + 1;
    ++index;
  }
  size_t distance;
  char code;

  // Loop.
  while (bIndex < bLength) {
    code = b[bIndex];
    result = distance = bIndex++;
    index = SIZE_MAX;

    while (++index < length) {
      bDistance = code == a[index] ? distance : distance + 1;
      distance = cache[index];

      cache[index] = result = distance > result ? bDistance > result
          ? result + 1
          : bDistance
          : bDistance > distance ? distance + 1 : bDistance;
    }
  }
  // printf("\nDistance from [%s] and [%s] is: [%zu]", a, b, result);
  free(cache);

  return result;
}

int levenshtein(const char *a, const char *b) {
  const size_t length = strlen(a);
  const size_t bLength = strlen(b);

  return (int)levenshtein_n(a, length, b, bLength);
}

// C helper functions:

void freeCharArray(char **a, int size) {
  int i;
  for (i = 0; i < size; i++) {
    if (a[i] != NULL) {
      printf("\nFreeing [%s]...", a[i]);
      free(a[i]);
    } else {
      printf("\n Unable to free!");
    }
  }
  free(a);
}

char *apply_levian_to_array(char **a, char *wrong, int size) {
  int i;
  int *ret_arr = malloc(sizeof(int) * size);
  char *correct = NULL;
  //printf("\napply_levian_to_array | START | Lenght: %d", size);
  for (i = 0; i < size; i++) {
    ret_arr[i] = levenshtein(a[i], wrong);
    if (ret_arr[i] == 1) {
      //printf("\nFOUND: [%s]", a[i]);
      free(ret_arr);
      return a[i];
    }
  }

  char alredy_found = '0';
  //printf("\n TRESHOLD: 2");
  for (i = 0; i < size; i++) {
    if (ret_arr[i] == 2) {
      //printf("\nFOUND: [%s]", a[i]);
      correct = a[i];
      alredy_found = '1';
      break;
    }
  }
  if (alredy_found == '0') {
    //printf("\n TRESHOLD: 3");
    for (i = 0; i < size; i++) {
      if (ret_arr[i] == 3) {
        //printf("\nFOUND: [%s]", a[i]);
        correct = a[i];
        alredy_found = '1';
        break;
      }
    }
  }
  if (alredy_found == '0') {
    //printf("\n TRESHOLD: 4");
    for (i = 0; i < size; i++) {
      if (ret_arr[i] == 4) {
        //printf("\nFOUND: [%s]", a[i]);
        correct = a[i];
        alredy_found = '1';
        break;
      }
    }
  }
 // printf("\napply_levian_to_array | STOP\n");

  free(ret_arr);
  return correct;
}

static void printCharArray(char **a, int size) {
  int i;
  printf("\nprintCharArray | START | Lenght: %d", size);
  for (i = 0; i < size; i++) {
    printf("\n%d) %s", i, (a[i]));
  }
  printf("\nprintCharArray | STOP\n");
  return;
}

char **split_data(char *str, int lenght) {

  char *saveptr;
  char **array = malloc(sizeof(char *) * lenght);
  char delim[] = "\n";
  int i = 0;
  char *ptr = strtok_r(str, delim, &saveptr);
  // printf("'%s'\n", ptr);
  array[i] = ptr;
  while ((ptr = strtok_r(NULL, delim, &saveptr)) != NULL) {
    array[++i] = ptr;
    //    printf("'%s'\n", ptr);
  }
  free(ptr);
//  printf("\n");
  return array;
}


char *spell_check(const char *filename, const int file_lines, char *wrong) {
  //printf("\nSTART | spell_check");
  char *dict = read_content(filename);

  char **array = split_data(dict, file_lines);
  // printCharArray(array,102264);
  char *correct = apply_levian_to_array(array, wrong, file_lines);
  printf("\nCorrect substitution for [%s] is [%s]\n", wrong, correct);

  //free(dict);
  free(array);
  //printf("\nSTOP | spell_check");
  return correct;
}

/*
int main() {
  const int dict_size = 102264;
  char *wrong = "zeccchin";
  char *right =  spell_check("all",dict_size,wrong);
  printf("\nSpell: %s",right);
  return 0;
}
*/

/*int main(int argc, char **argv)
  {
  char *content;
  printf("%ld",get_file_size("/tmp/test1"));
  content = read_content("filename.txt");
  verify_presence_data(content,"provaT");
  verify_presence_filename(content,"filename.txt");
  free(content);
  return 0;
  }*/
